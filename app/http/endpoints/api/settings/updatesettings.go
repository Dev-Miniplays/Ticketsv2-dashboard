package api

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/Miniplays-Tickets/dashboard/botcontext"
	dbclient "github.com/Miniplays-Tickets/dashboard/database"
	"github.com/Miniplays-Tickets/dashboard/rpc"
	"github.com/Miniplays-Tickets/dashboard/rpc/cache"
	"github.com/Miniplays-Tickets/dashboard/utils"
	"github.com/Miniplays-Tickets/worker/bot/customisation"
	"github.com/Miniplays-Tickets/worker/i18n"
	"github.com/TicketsBot-cloud/common/premium"
	"github.com/TicketsBot-cloud/database"
	"github.com/gin-gonic/gin"
	"github.com/rxdn/gdl/objects/channel"
	"golang.org/x/sync/errgroup"
)

func UpdateSettingsHandler(ctx *gin.Context) {
	guildId := ctx.Keys["guildid"].(uint64)

	var settings Settings
	if err := ctx.BindJSON(&settings); err != nil {
		ctx.JSON(400, utils.ErrorJson(err))
		return
	}

	// Get a list of all channel IDs
	botContext, err := botcontext.ContextForGuild(guildId)
	if err != nil {
		ctx.JSON(500, utils.ErrorJson(err))
		return
	}

	// TODO: Use proper context
	channels, err := botContext.GetGuildChannels(context.Background(), guildId)
	if err != nil {
		ctx.JSON(500, utils.ErrorJson(err))
		return
	}

	// Includes voting
	premiumTier, err := rpc.PremiumClient.GetTierByGuildId(ctx, guildId, true, botContext.Token, botContext.RateLimiter)
	if err != nil {
		ctx.JSON(500, utils.ErrorJson(err))
		return
	}

	if err := settings.Validate(ctx, guildId, premiumTier); err != nil {
		ctx.JSON(400, utils.ErrorJson(err))
		return
	}

	group, _ := errgroup.WithContext(context.Background())

	group.Go(func() error {
		return settings.updateSettings(ctx, guildId)
	})

	group.Go(func() error {
		return settings.updateClaimSettings(ctx, guildId)
	})

	addToWaitGroup(group, guildId, settings.updateTicketPermissions)
	addToWaitGroup(group, guildId, settings.updateLanguage)
	addToWaitGroup(group, guildId, settings.updateAutoClose)

	if premiumTier > premium.None {
		addToWaitGroup(group, guildId, settings.updateColours)
	}

	// TODO: Errors
	var errStr *string = nil
	if err := group.Wait(); err != nil {
		errStr = utils.Ptr(err.Error())
	}

	validWelcomeMessage := settings.updateWelcomeMessage(guildId)
	validTicketLimit := settings.updateTicketLimit(guildId)
	validArchiveChannel := settings.updateArchiveChannel(channels, guildId)
	validCategory := settings.updateCategory(channels, guildId)
	validNamingScheme := settings.updateNamingScheme(guildId)
	settings.updateUsersCanClose(guildId)
	settings.updateCloseConfirmation(guildId)
	settings.updateFeedbackEnabled(guildId)

	ctx.JSON(200, gin.H{
		"welcome_message": validWelcomeMessage,
		"ticket_limit":    validTicketLimit,
		"archive_channel": validArchiveChannel,
		"category":        validCategory,
		"naming_scheme":   validNamingScheme,
		"error":           errStr,
	})
}

func (s *Settings) updateSettings(ctx context.Context, guildId uint64) error {
	return dbclient.Client.Settings.Set(ctx, guildId, s.Settings)
}

func (s *Settings) updateClaimSettings(ctx context.Context, guildId uint64) error {
	return dbclient.Client.ClaimSettings.Set(ctx, guildId, s.ClaimSettings)
}

var (
	validAutoArchive = []int{60, 1440, 4320, 10080}
	activeColours    = []customisation.Colour{customisation.Green, customisation.Red}
)

func (s *Settings) Validate(ctx context.Context, guildId uint64, premiumTier premium.PremiumTier) error {
	// Sync checks
	if s.ClaimSettings.SupportCanType && !s.ClaimSettings.SupportCanView {
		return errors.New("Du musst den Kanal sehen können, um tippen zu können")
	}

	if s.Settings.UseThreads && s.TicketNotificationChannel == nil {
		return errors.New("Du musst einen Ticket Benachrigungskanal auswählen")
	}

	if !s.Settings.UseThreads {
		s.TicketNotificationChannel = nil
	}

	if s.Language != nil {
		if _, ok := i18n.MappedByIsoShortCode[*s.Language]; !ok {
			return errors.New("Ungültige Sprache")
		}
	}

	// Validate colours
	if len(s.Colours) > len(activeColours) {
		return errors.New("Ungültige Farbe")
	}

	for colour := range s.Colours {
		if !utils.Exists(activeColours, colour) {
			return errors.New("Ungültige Farbe")
		}
	}

	for _, colourCode := range activeColours {
		if _, ok := s.Colours[colourCode]; !ok {
			s.Colours[colourCode] = utils.HexColour(customisation.DefaultColours[colourCode])
		}
	}

	// Validate autoclose
	if premiumTier < premium.Premium {
		s.AutoCloseSettings.SinceOpenWithNoResponse = 0
		s.AutoCloseSettings.SinceLastMessage = 0
	}

	if !s.AutoCloseSettings.Enabled {
		s.AutoCloseSettings.SinceOpenWithNoResponse = 0
		s.AutoCloseSettings.SinceLastMessage = 0
		s.AutoCloseSettings.OnUserLeave = false
	}

	if s.AutoCloseSettings.SinceOpenWithNoResponse < 0 {
		s.AutoCloseSettings.SinceOpenWithNoResponse = 0
	}

	if s.AutoCloseSettings.SinceLastMessage < 0 {
		s.AutoCloseSettings.SinceLastMessage = 0
	}

	if s.AutoCloseSettings.SinceLastMessage > int64((time.Hour*24*60).Seconds()) ||
		s.AutoCloseSettings.SinceOpenWithNoResponse > int64((time.Hour*24*60).Seconds()) {
		return errors.New("Der Zeitraum für das automatische Schließen darf 60 Tage nicht überschreiten")
	}

	// Async checks
	group, _ := errgroup.WithContext(context.Background())

	// Validate panel from same guild
	group.Go(func() error {
		if s.ContextMenuPanel != nil {
			panelId := *s.ContextMenuPanel

			panel, err := dbclient.Client.Panel.GetById(ctx, panelId)
			if err != nil {
				return err
			}

			if guildId != panel.GuildId {
				return fmt.Errorf("Guild ID stimmt nicht")
			}
		}

		return nil
	})

	group.Go(func() error {
		valid := false
		for _, duration := range validAutoArchive {
			if duration == s.Settings.ThreadArchiveDuration {
				valid = true
				break
			}
		}

		if !valid {
			return fmt.Errorf("Ungültige Dauer für das automatische Archivieren des Threads")
		}

		return nil
	})

	group.Go(func() error {
		if s.Settings.OverflowCategoryId != nil {
			ctx, cancel := context.WithTimeout(context.Background(), time.Second*5)
			defer cancel()

			ch, err := cache.Instance.GetChannel(ctx, *s.Settings.OverflowCategoryId)
			if err != nil {
				return fmt.Errorf("Ungültige Überlauf Kategorie")
			}

			if ch.GuildId != guildId {
				return fmt.Errorf("Überlauf Kategorie Guild ID stimmt nicht")
			}

			if ch.Type != channel.ChannelTypeGuildCategory {
				return fmt.Errorf("Überlauf Kategorie ist keine Kategorie")
			}
		}

		return nil
	})

	group.Go(func() error {
		if s.Settings.TicketNotificationChannel != nil {
			ctx, cancel := context.WithTimeout(context.Background(), time.Second*5)
			defer cancel()

			ch, err := cache.Instance.GetChannel(ctx, *s.Settings.TicketNotificationChannel)
			if err != nil {
				return fmt.Errorf("Ungültiger Ticket Benachrichtigungskanal")
			}

			if ch.GuildId != guildId {
				return fmt.Errorf("Ticket Benachrichtigungskanall Guild ID stimmt nicht")
			}

			if ch.Type != channel.ChannelTypeGuildText {
				return fmt.Errorf("Ticket Benachrichtigungskanal ist kein Text Kanal")
			}
		}

		return nil
	})

	return group.Wait()
}

func addToWaitGroup(group *errgroup.Group, guildId uint64, f func(uint64) error) {
	group.Go(func() error {
		return f(guildId)
	})
}

func (s *Settings) updateWelcomeMessage(guildId uint64) bool {
	if s.WelcomeMessage == "" || len(s.WelcomeMessage) > 4096 {
		return false
	}

	go dbclient.Client.WelcomeMessages.Set(context.Background(), guildId, s.WelcomeMessage)
	return true
}

func (s *Settings) updateTicketLimit(guildId uint64) bool {
	if s.TicketLimit > 10 || s.TicketLimit < 1 {
		return false
	}

	go dbclient.Client.TicketLimit.Set(context.Background(), guildId, s.TicketLimit)
	return true
}

func (s *Settings) updateCategory(channels []channel.Channel, guildId uint64) bool {
	var valid bool
	for _, ch := range channels {
		if ch.Id == s.Category && ch.Type == channel.ChannelTypeGuildCategory {
			valid = true
			break
		}
	}

	if !valid {
		return false
	}

	go dbclient.Client.ChannelCategory.Set(context.Background(), guildId, s.Category)
	return true
}

func (s *Settings) updateArchiveChannel(channels []channel.Channel, guildId uint64) bool {
	if s.ArchiveChannel == nil {
		go dbclient.Client.ArchiveChannel.Set(context.Background(), guildId, nil)
		return true
	}

	var valid bool
	for _, ch := range channels {
		if ch.Id == *s.ArchiveChannel && ch.Type == channel.ChannelTypeGuildText {
			valid = true
			break
		}
	}

	if !valid {
		return false
	}

	go dbclient.Client.ArchiveChannel.Set(context.Background(), guildId, s.ArchiveChannel)
	return true
}

var validScheme = []database.NamingScheme{database.Id, database.Username}

func (s *Settings) updateNamingScheme(guildId uint64) bool {
	var valid bool
	for _, scheme := range validScheme {
		if scheme == s.NamingScheme {
			valid = true
			break
		}
	}

	if !valid {
		return false
	}

	go dbclient.Client.NamingScheme.Set(context.Background(), guildId, s.NamingScheme)
	return true
}

func (s *Settings) updateUsersCanClose(guildId uint64) {
	go dbclient.Client.UsersCanClose.Set(context.Background(), guildId, s.UsersCanClose)
}

func (s *Settings) updateCloseConfirmation(guildId uint64) {
	go dbclient.Client.CloseConfirmation.Set(context.Background(), guildId, s.CloseConfirmation)
}

func (s *Settings) updateFeedbackEnabled(guildId uint64) {
	go dbclient.Client.FeedbackEnabled.Set(context.Background(), guildId, s.FeedbackEnabled)
}

func (s *Settings) updateLanguage(guildId uint64) error {
	if s.Language == nil {
		return dbclient.Client.ActiveLanguage.Delete(context.Background(), guildId)
	} else {
		return dbclient.Client.ActiveLanguage.Set(context.Background(), guildId, string(*s.Language))
	}
}

func (s *Settings) updateTicketPermissions(guildId uint64) error {
	return dbclient.Client.TicketPermissions.Set(context.Background(), guildId, s.TicketPermissions) // No validation required
}

func (s *Settings) updateColours(guildId uint64) error {
	// Convert ColourMap to primitives
	converted := make(map[int16]int)
	for colour, hex := range s.Colours {
		converted[int16(colour)] = int(hex)
	}

	return dbclient.Client.CustomColours.BatchSet(context.Background(), guildId, converted)
}

func (s *Settings) updateAutoClose(guildId uint64) error {
	data := s.AutoCloseSettings.ConvertToDatabase() // Already validated
	return dbclient.Client.AutoClose.Set(context.Background(), guildId, data)
}

func (d AutoCloseData) ConvertToDatabase() (settings database.AutoCloseSettings) {
	settings.Enabled = d.Enabled

	if d.SinceOpenWithNoResponse > 0 {
		duration := time.Second * time.Duration(d.SinceOpenWithNoResponse)
		settings.SinceOpenWithNoResponse = &duration
	}

	if d.SinceLastMessage > 0 {
		duration := time.Second * time.Duration(d.SinceLastMessage)
		settings.SinceLastMessage = &duration
	}

	settings.OnUserLeave = &d.OnUserLeave
	return
}
