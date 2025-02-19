package api

import (
	"github.com/Dev-Miniplays/Ticketsv2-dashboard/botcontext"
	"github.com/Dev-Miniplays/Ticketsv2-dashboard/utils"
	"github.com/gin-gonic/gin"
)

func RolesHandler(ctx *gin.Context) {
	guildId := ctx.Keys["guildid"].(uint64)

	botContext, err := botcontext.ContextForGuild(guildId)
	if err != nil {
		ctx.JSON(500, utils.ErrorJson(err))
		return
	}

	roles, err := botContext.RestCache.GetGuildRoles(guildId)
	if err != nil {
		ctx.JSON(500, utils.ErrorJson(err))
		return
	}

	ctx.JSON(200, gin.H{
		"success": true,
		"roles":   roles,
	})
}
