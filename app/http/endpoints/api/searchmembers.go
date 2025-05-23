package api

import (
	"context"

	"github.com/Miniplays-Tickets/dashboard/botcontext"
	"github.com/Miniplays-Tickets/dashboard/utils"
	"github.com/gin-gonic/gin"
	"github.com/rxdn/gdl/objects/member"
)

func SearchMembers(ctx *gin.Context) {
	guildId := ctx.Keys["guildid"].(uint64)

	botCtx, err := botcontext.ContextForGuild(guildId)
	if err != nil {
		ctx.JSON(500, utils.ErrorJson(err))
		return
	}

	query := ctx.Query("query")
	if len(query) > 32 {
		ctx.JSON(400, utils.ErrorStr("Ungültige query"))
		return
	}

	var members []member.Member
	if query == "" {
		// TODO: Use proper context
		members, err = botCtx.ListMembers(context.Background(), guildId)
	} else {
		// TODO: Use proper context
		members, err = botCtx.SearchMembers(context.Background(), guildId, query)
	}

	if err != nil {
		ctx.JSON(500, utils.ErrorJson(err))
		return
	}

	ctx.JSON(200, members)
}
