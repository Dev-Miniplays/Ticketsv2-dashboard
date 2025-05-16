package api

import (
	"github.com/Miniplays-Tickets/dashboard/database"
	"github.com/Miniplays-Tickets/dashboard/utils"
	"github.com/gin-gonic/gin"
)

func GetOverrideHandler(ctx *gin.Context) {
	guildId := ctx.Keys["guildid"].(uint64)

	hasOverride, err := database.Client.StaffOverride.HasActiveOverride(ctx, guildId)
	if err != nil {
		ctx.JSON(500, utils.ErrorJson(err))
		return
	}

	ctx.JSON(200, gin.H{
		"has_override": hasOverride,
	})
}
