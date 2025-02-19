package api

import (
	"github.com/Dev-Miniplays/Ticketsv2-dashboard/database"
	"github.com/Dev-Miniplays/Ticketsv2-dashboard/utils"
	"github.com/Dev-Miniplays/Ticketsv2-dashboard/utils/types"
	"github.com/gin-gonic/gin"
)

func TagsListHandler(ctx *gin.Context) {
	guildId := ctx.Keys["guildid"].(uint64)

	tags, err := database.Client.Tag.GetByGuild(ctx, guildId)
	if err != nil {
		ctx.JSON(500, utils.ErrorJson(err))
		return
	}

	wrapped := make(map[string]tag)
	for id, data := range tags {
		var embed *types.CustomEmbed
		if data.Embed != nil {
			embed = types.NewCustomEmbed(data.Embed.CustomEmbed, data.Embed.Fields)
		}

		wrapped[id] = tag{
			Id:              data.Id,
			UseGuildCommand: data.ApplicationCommandId != nil,
			Content:         data.Content,
			UseEmbed:        data.Embed != nil,
			Embed:           embed,
		}
	}

	ctx.JSON(200, wrapped)
}
