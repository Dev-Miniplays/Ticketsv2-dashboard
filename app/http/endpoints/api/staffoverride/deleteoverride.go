package api

import (
	"github.com/Miniplays-Tickets/dashboard/database"
	"github.com/Miniplays-Tickets/dashboard/utils"
	"github.com/gin-gonic/gin"
)

func DeleteOverrideHandler(ctx *gin.Context) {
	guildId := ctx.Keys["guildid"].(uint64)

	if err := database.Client.StaffOverride.Delete(ctx, guildId); err != nil {
		ctx.JSON(500, utils.ErrorJson(err))
		return
	}

	ctx.Status(204)
}
