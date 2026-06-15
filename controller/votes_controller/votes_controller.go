package votescontroller

import (
	"errors"
	"strconv"
	"strings"

	"github.com/gofiber/fiber/v2"

	votes "trendflix/services/votes"
	"trendflix/utils/httpx"
)

func VotePost(svc *votes.Service) fiber.Handler {
	return func(c *fiber.Ctx) error {
		user, ok := httpx.CurrentUser(c)
		if !ok {
			return httpx.Error(c, fiber.StatusUnauthorized, "Unauthorized")
		}

		id, err := strconv.ParseUint(strings.TrimSpace(c.Params("id")), 10, 64)
		if err != nil || id == 0 {
			return httpx.Error(c, fiber.StatusBadRequest, "Invalid post id")
		}

		var input votes.VoteInput
		if err := c.BodyParser(&input); err != nil {
			return httpx.Error(c, fiber.StatusBadRequest, "Invalid request body")
		}

		post, err := svc.VotePost(user.ID, uint(id), input)
		if err != nil {
			return mapVoteError(c, err)
		}
		return httpx.SuccessData(c, fiber.StatusOK, "Vote registered", "post", post)
	}
}

func VoteComment(svc *votes.Service) fiber.Handler {
	return func(c *fiber.Ctx) error {
		user, ok := httpx.CurrentUser(c)
		if !ok {
			return httpx.Error(c, fiber.StatusUnauthorized, "Unauthorized")
		}

		id, err := strconv.ParseUint(strings.TrimSpace(c.Params("id")), 10, 64)
		if err != nil || id == 0 {
			return httpx.Error(c, fiber.StatusBadRequest, "Invalid comment id")
		}

		var input votes.VoteInput
		if err := c.BodyParser(&input); err != nil {
			return httpx.Error(c, fiber.StatusBadRequest, "Invalid request body")
		}

		comment, err := svc.VoteComment(user.ID, uint(id), input)
		if err != nil {
			return mapVoteError(c, err)
		}
		return httpx.SuccessData(c, fiber.StatusOK, "Vote registered", "comment", comment)
	}
}

func mapVoteError(c *fiber.Ctx, err error) error {
	switch {
	case errors.Is(err, votes.ErrNotFound):
		return httpx.Error(c, fiber.StatusNotFound, "Target not found")
	case errors.Is(err, votes.ErrInvalidVoteType):
		return httpx.ValidationError(c, "vote_type must be 'up' or 'down'", fiber.Map{"vote_type": "must be 'up' or 'down'"})
	default:
		return httpx.Error(c, fiber.StatusInternalServerError, "Something went wrong")
	}
}
