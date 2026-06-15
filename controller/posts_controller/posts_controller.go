package postscontroller

import (
	"errors"
	"strconv"
	"strings"

	"github.com/gofiber/fiber/v2"

	communities "trendflix/services/communities"
	posts "trendflix/services/posts"
	"trendflix/utils/httpx"
)

func ListByCommunity(svc *posts.Service) fiber.Handler {
	return func(c *fiber.Ctx) error {
		communityID, err := strconv.ParseUint(strings.TrimSpace(c.Params("id")), 10, 64)
		if err != nil || communityID == 0 {
			return httpx.Error(c, fiber.StatusBadRequest, "Invalid community id")
		}

		sort := strings.TrimSpace(c.Query("sort"))
		page, _ := strconv.Atoi(strings.TrimSpace(c.Query("page")))
		perPage, _ := strconv.Atoi(strings.TrimSpace(c.Query("per_page")))

		result, err := svc.ListByCommunity(uint(communityID), sort, page, perPage)
		if err != nil {
			return httpx.Error(c, fiber.StatusInternalServerError, "Failed to fetch posts")
		}
		return httpx.Success(c, fiber.StatusOK, "Posts fetched successfully", fiber.Map{
			"posts": result.Items,
			"total": result.Total,
			"page":  result.Page,
			"pages": result.Pages,
		})
	}
}

func ListByItem(svc *posts.Service) fiber.Handler {
	return func(c *fiber.Ctx) error {
		itemType := strings.TrimSpace(c.Params("type"))
		itemID, err := strconv.ParseUint(strings.TrimSpace(c.Params("id")), 10, 64)
		if err != nil || itemID == 0 {
			return httpx.Error(c, fiber.StatusBadRequest, "Invalid item id")
		}

		page, _ := strconv.Atoi(strings.TrimSpace(c.Query("page")))
		perPage, _ := strconv.Atoi(strings.TrimSpace(c.Query("per_page")))

		result, err := svc.ListByItem(itemType, uint(itemID), page, perPage)
		if err != nil {
			return httpx.Error(c, fiber.StatusInternalServerError, "Failed to fetch discussions")
		}
		return httpx.Success(c, fiber.StatusOK, "Discussions fetched successfully", fiber.Map{
			"posts": result.Items,
			"total": result.Total,
			"page":  result.Page,
			"pages": result.Pages,
		})
	}
}

func GetByID(svc *posts.Service) fiber.Handler {
	return func(c *fiber.Ctx) error {
		id, err := strconv.ParseUint(strings.TrimSpace(c.Params("id")), 10, 64)
		if err != nil || id == 0 {
			return httpx.Error(c, fiber.StatusBadRequest, "Invalid post id")
		}

		post, err := svc.GetByID(uint(id))
		if err != nil {
			if errors.Is(err, posts.ErrNotFound) {
				return httpx.Error(c, fiber.StatusNotFound, "Post not found")
			}
			return httpx.Error(c, fiber.StatusInternalServerError, "Failed to fetch post")
		}
		return httpx.SuccessData(c, fiber.StatusOK, "Post fetched successfully", "post", post)
	}
}

func Create(svc *posts.Service) fiber.Handler {
	return func(c *fiber.Ctx) error {
		user, ok := httpx.CurrentUser(c)
		if !ok {
			return httpx.Error(c, fiber.StatusUnauthorized, "Unauthorized")
		}

		communityID, err := strconv.ParseUint(strings.TrimSpace(c.Params("id")), 10, 64)
		if err != nil || communityID == 0 {
			return httpx.Error(c, fiber.StatusBadRequest, "Invalid community id")
		}

		var input posts.CreateInput
		if err := c.BodyParser(&input); err != nil {
			return httpx.Error(c, fiber.StatusBadRequest, "Invalid request body")
		}

		post, err := svc.Create(user.ID, uint(communityID), input)
		if err != nil {
			return mapPostError(c, err)
		}
		return httpx.SuccessData(c, fiber.StatusCreated, "Post created successfully", "post", post)
	}
}

func Update(svc *posts.Service) fiber.Handler {
	return func(c *fiber.Ctx) error {
		user, ok := httpx.CurrentUser(c)
		if !ok {
			return httpx.Error(c, fiber.StatusUnauthorized, "Unauthorized")
		}

		id, err := strconv.ParseUint(strings.TrimSpace(c.Params("id")), 10, 64)
		if err != nil || id == 0 {
			return httpx.Error(c, fiber.StatusBadRequest, "Invalid post id")
		}

		var input posts.UpdateInput
		if err := c.BodyParser(&input); err != nil {
			return httpx.Error(c, fiber.StatusBadRequest, "Invalid request body")
		}

		post, err := svc.Update(user.ID, uint(id), input)
		if err != nil {
			return mapPostError(c, err)
		}
		return httpx.SuccessData(c, fiber.StatusOK, "Post updated successfully", "post", post)
	}
}

func Delete(svc *posts.Service) fiber.Handler {
	return func(c *fiber.Ctx) error {
		user, ok := httpx.CurrentUser(c)
		if !ok {
			return httpx.Error(c, fiber.StatusUnauthorized, "Unauthorized")
		}

		id, err := strconv.ParseUint(strings.TrimSpace(c.Params("id")), 10, 64)
		if err != nil || id == 0 {
			return httpx.Error(c, fiber.StatusBadRequest, "Invalid post id")
		}

		if err := svc.Delete(user.ID, uint(id)); err != nil {
			return mapPostError(c, err)
		}
		return httpx.Success(c, fiber.StatusOK, "Post deleted successfully", nil)
	}
}

func mapPostError(c *fiber.Ctx, err error) error {
	switch {
	case errors.Is(err, posts.ErrNotFound), errors.Is(err, posts.ErrCommunityNotFound):
		return httpx.Error(c, fiber.StatusNotFound, err.Error())
	case errors.Is(err, posts.ErrInvalidTitle):
		return httpx.ValidationError(c, "Title is required", fiber.Map{"title": "title is required"})
	case errors.Is(err, posts.ErrMustJoin):
		return httpx.Error(c, fiber.StatusForbidden, "You must join this community before posting")
	case errors.Is(err, posts.ErrBanned):
		return httpx.Error(c, fiber.StatusForbidden, "You are banned from this community")
	case errors.Is(err, posts.ErrForbidden):
		return httpx.Error(c, fiber.StatusForbidden, "You are not allowed to perform this action")
	case errors.Is(err, posts.ErrRateLimitPost):
		return httpx.Error(c, fiber.StatusTooManyRequests, "You are posting too fast, try again later")
	case errors.Is(err, posts.ErrInvalidPostType), errors.Is(err, posts.ErrInvalidItemType), errors.Is(err, posts.ErrItemIDRequired):
		return httpx.ValidationError(c, err.Error(), fiber.Map{})
	case errors.Is(err, communities.ErrNotFound):
		return httpx.Error(c, fiber.StatusNotFound, "Community not found")
	default:
		return httpx.Error(c, fiber.StatusInternalServerError, "Something went wrong")
	}
}
