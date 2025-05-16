package api

import (
	"net/http"

	"github.com/Dev-Miniplays/Ticketsv2-dashboard/app"
	"github.com/Dev-Miniplays/Ticketsv2-dashboard/database"
	"github.com/Dev-Miniplays/Ticketsv2-dashboard/redis"
	"github.com/Dev-Miniplays/Ticketsv2-dashboard/utils"
	"github.com/TicketsBot-cloud/common/statusupdates"
	"github.com/gin-gonic/gin"
	"github.com/rxdn/gdl/objects/user"
)

type statusUpdateBody struct {
	Status     string            `json:"status"`
	StatusType user.ActivityType `json:"status_type,string"`
}

func WhitelabelStatusPost(c *gin.Context) {
	userId := c.Keys["userid"].(uint64)

	// Get bot
	bot, err := database.Client.Whitelabel.GetByUserId(c, userId)
	if err != nil {
		_ = c.AbortWithError(http.StatusInternalServerError, app.NewServerError(err))
		return
	}

	// Ensure bot exists
	if bot.BotId == 0 {
		c.JSON(404, utils.ErrorStr("Kein Bot gefunden"))
		return
	}

	// Parse status
	var data statusUpdateBody
	if err := c.BindJSON(&data); err != nil {
		c.JSON(400, utils.ErrorStr("Fehler 11"))
		return
	}

	// Validate status length
	if len(data.Status) == 0 || len(data.Status) > 255 {
		c.JSON(400, utils.ErrorStr("Status muss zwischen 1-255 Zeichen lang sein"))
		return
	}

	// Validate status type
	validActivities := []user.ActivityType{
		user.ActivityTypePlaying,
		user.ActivityTypeListening,
		user.ActivityTypeWatching,
	}

	if !utils.Contains(validActivities, data.StatusType) {
		c.JSON(400, utils.ErrorStr("Ungültiger Status Typ"))
		return
	}

	// Update in database
	if err := database.Client.WhitelabelStatuses.Set(c, bot.BotId, data.Status, int16(data.StatusType)); err != nil {
		_ = c.AbortWithError(http.StatusInternalServerError, app.NewServerError(err))
		return
	}

	// Send status update to sharder
	go statusupdates.Publish(redis.Client.Client, bot.BotId)

	c.JSON(200, utils.SuccessResponse)
}
