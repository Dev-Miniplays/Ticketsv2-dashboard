package api

import (
	"context"
	"errors"
	"strconv"

	"github.com/Miniplays-Tickets/dashboard/botcontext"
	dbclient "github.com/Miniplays-Tickets/dashboard/database"
	"github.com/Miniplays-Tickets/dashboard/rpc"
	"github.com/Miniplays-Tickets/dashboard/utils"
	"github.com/TicketsBot-cloud/common/premium"
	"github.com/gin-gonic/gin"
	"github.com/rxdn/gdl/rest"
	"github.com/rxdn/gdl/rest/request"
)

func MultiPanelResend(ctx *gin.Context) {
	guildId := ctx.Keys["guildid"].(uint64)

	// parse panel ID
	panelId, err := strconv.Atoi(ctx.Param("panelid"))
	if err != nil {
		ctx.JSON(400, utils.ErrorJson(err))
		return
	}

	// retrieve panel from DB
	multiPanel, ok, err := dbclient.Client.MultiPanels.Get(ctx, panelId)
	if err != nil {
		ctx.JSON(500, utils.ErrorJson(err))
		return
	}

	// check panel exists
	if !ok {
		ctx.JSON(404, utils.ErrorStr("Kein Panel mit der angegebenen ID gefunden"))
		return
	}

	// check panel is in the same guild
	if guildId != multiPanel.GuildId {
		ctx.JSON(403, utils.ErrorStr("Guild ID stimmt nicht"))
		return
	}

	// get bot context
	botContext, err := botcontext.ContextForGuild(guildId)
	if err != nil {
		ctx.JSON(500, utils.ErrorJson(err))
		return
	}

	// delete old message
	// TODO: Use proper context
	if err := rest.DeleteMessage(context.Background(), botContext.Token, botContext.RateLimiter, multiPanel.ChannelId, multiPanel.MessageId); err != nil {
		var unwrapped request.RestError
		if errors.As(err, &unwrapped) && !unwrapped.IsClientError() {
			ctx.JSON(500, utils.ErrorJson(err))
			return
		}
	}

	// get premium status
	premiumTier, err := rpc.PremiumClient.GetTierByGuildId(ctx, guildId, true, botContext.Token, botContext.RateLimiter)
	if err != nil {
		ctx.JSON(500, utils.ErrorJson(err))
		return
	}

	panels, err := dbclient.Client.MultiPanelTargets.GetPanels(ctx, multiPanel.Id)
	if err != nil {
		ctx.JSON(500, utils.ErrorJson(err))
		return
	}

	// send new message
	messageData := multiPanelIntoMessageData(multiPanel, premiumTier > premium.None)
	messageId, err := messageData.send(botContext, panels)
	if err != nil {
		var unwrapped request.RestError
		if errors.As(err, &unwrapped) && unwrapped.StatusCode == 403 {
			ctx.JSON(500, utils.ErrorJson(errors.New("Ich habe keine Berechtigung, Nachrichten in dem angegebenen Kanal zu senden.")))
		} else {
			ctx.JSON(500, utils.ErrorJson(err))
		}

		return
	}

	if err = dbclient.Client.MultiPanels.UpdateMessageId(ctx, multiPanel.Id, messageId); err != nil {
		ctx.JSON(500, utils.ErrorJson(err))
		return
	}

	ctx.JSON(200, gin.H{
		"success": true,
	})
}
