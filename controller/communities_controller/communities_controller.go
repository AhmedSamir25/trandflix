package communitiescontroller

import (
	"errors"
	"strconv"
	"strings"

	"github.com/gofiber/fiber/v2"

	communities "trendflix/services/communities"
	"trendflix/utils/httpx"
)

func parsePage(c *fiber.Ctx) (int, int) {
	page, _ := strconv.Atoi(strings.TrimSpace(c.Query("page")))
	perPage, _ := strconv.Atoi(strings.TrimSpace(c.Query("per_page")))
	return page, perPage
}

func List(svc *communities.Service) fiber.Handler {
	return func(c *fiber.Ctx) error {
		search := strings.TrimSpace(c.Query("q"))
		page, perPage := parsePage(c)

		var userID uint
		if user, ok := httpx.CurrentUser(c); ok {
			userID = user.ID
		}

		result, err := svc.List(search, page, perPage, userID)
		if err != nil {
			return httpx.Error(c, fiber.StatusInternalServerError, "Failed to fetch communities")
		}
		return httpx.Success(c, fiber.StatusOK, "Communities fetched successfully", fiber.Map{
			"communities": result.Items,
			"total":       result.Total,
			"page":        result.Page,
			"pages":       result.Pages,
		})
	}
}

func AdminList(svc *communities.Service) fiber.Handler {
	return func(c *fiber.Ctx) error {
		search := strings.TrimSpace(c.Query("q"))
		status := strings.TrimSpace(c.Query("status"))
		categoryType := strings.TrimSpace(c.Query("category_type"))
		page, perPage := parsePage(c)

		result, err := svc.ListAll(search, status, categoryType, page, perPage)
		if err != nil {
			return httpx.Error(c, fiber.StatusInternalServerError, "Failed to fetch communities")
		}
		return httpx.Success(c, fiber.StatusOK, "Communities fetched successfully", fiber.Map{
			"communities": result.Items,
			"total":       result.Total,
			"page":        result.Page,
			"pages":       result.Pages,
		})
	}
}

func Stats(svc *communities.Service) fiber.Handler {
	return func(c *fiber.Ctx) error {
		stats, err := svc.Stats()
		if err != nil {
			return httpx.Error(c, fiber.StatusInternalServerError, "Failed to fetch community stats")
		}
		return httpx.SuccessData(c, fiber.StatusOK, "Community stats fetched successfully", "stats", stats)
	}
}

func AdminDelete(svc *communities.Service) fiber.Handler {
	return func(c *fiber.Ctx) error {
		id, err := strconv.ParseUint(strings.TrimSpace(c.Params("id")), 10, 64)
		if err != nil || id == 0 {
			return httpx.Error(c, fiber.StatusBadRequest, "Invalid community id")
		}

		if err := svc.AdminDelete(uint(id)); err != nil {
			return mapCommunityError(c, err)
		}
		return httpx.Success(c, fiber.StatusOK, "Community deleted successfully", nil)
	}
}

func Block(svc *communities.Service) fiber.Handler {
	return func(c *fiber.Ctx) error {
		return setAdminStatus(c, svc, communities.StatusRejected)
	}
}

func Unblock(svc *communities.Service) fiber.Handler {
	return func(c *fiber.Ctx) error {
		return setAdminStatus(c, svc, communities.StatusApproved)
	}
}

func setAdminStatus(c *fiber.Ctx, svc *communities.Service, status string) error {
	id, err := strconv.ParseUint(strings.TrimSpace(c.Params("id")), 10, 64)
	if err != nil || id == 0 {
		return httpx.Error(c, fiber.StatusBadRequest, "Invalid community id")
	}

	community, err := svc.SetStatus(uint(id), status)
	if err != nil {
		return mapCommunityError(c, err)
	}
	return httpx.SuccessData(c, fiber.StatusOK, "Community status updated", "community", community)
}

func Popular(svc *communities.Service) fiber.Handler {
	return func(c *fiber.Ctx) error {
		limit, _ := strconv.Atoi(strings.TrimSpace(c.Query("limit")))
		items, err := svc.Popular(limit)
		if err != nil {
			return httpx.Error(c, fiber.StatusInternalServerError, "Failed to fetch popular communities")
		}
		return httpx.Success(c, fiber.StatusOK, "Popular communities fetched", fiber.Map{
			"communities": items,
		})
	}
}

func Recommended(svc *communities.Service) fiber.Handler {
	return func(c *fiber.Ctx) error {
		user, ok := httpx.CurrentUser(c)
		if !ok {
			items, err := svc.Popular(8)
			if err != nil {
				return httpx.Error(c, fiber.StatusInternalServerError, "Failed to fetch recommended communities")
			}
			return httpx.Success(c, fiber.StatusOK, "Recommended communities fetched", fiber.Map{
				"communities": items,
			})
		}
		_ = user
		items, err := svc.Popular(8)
		if err != nil {
			return httpx.Error(c, fiber.StatusInternalServerError, "Failed to fetch recommended communities")
		}
		return httpx.Success(c, fiber.StatusOK, "Recommended communities fetched", fiber.Map{
			"communities": items,
		})
	}
}

func GetBySlug(svc *communities.Service) fiber.Handler {
	return func(c *fiber.Ctx) error {
		slug := strings.TrimSpace(c.Params("slug"))
		if slug == "" {
			return httpx.Error(c, fiber.StatusBadRequest, "Invalid slug")
		}
		community, err := svc.GetBySlug(slug)
		if err != nil {
			if errors.Is(err, communities.ErrNotFound) {
				return httpx.Error(c, fiber.StatusNotFound, "Community not found")
			}
			return httpx.Error(c, fiber.StatusInternalServerError, "Failed to fetch community")
		}
		return httpx.SuccessData(c, fiber.StatusOK, "Community fetched successfully", "community", community)
	}
}

func Create(svc *communities.Service) fiber.Handler {
	return func(c *fiber.Ctx) error {
		user, ok := httpx.CurrentUser(c)
		if !ok {
			return httpx.Error(c, fiber.StatusUnauthorized, "Unauthorized")
		}

		var input communities.CreateInput
		if err := c.BodyParser(&input); err != nil {
			return httpx.Error(c, fiber.StatusBadRequest, "Invalid request body")
		}

		community, err := svc.Create(user.ID, input)
		if err != nil {
			return mapCommunityError(c, err)
		}
		return httpx.SuccessData(c, fiber.StatusCreated, "Community created successfully", "community", community)
	}
}

func Update(svc *communities.Service) fiber.Handler {
	return func(c *fiber.Ctx) error {
		user, ok := httpx.CurrentUser(c)
		if !ok {
			return httpx.Error(c, fiber.StatusUnauthorized, "Unauthorized")
		}

		id, err := strconv.ParseUint(strings.TrimSpace(c.Params("id")), 10, 64)
		if err != nil || id == 0 {
			return httpx.Error(c, fiber.StatusBadRequest, "Invalid community id")
		}

		var input communities.UpdateInput
		if err := c.BodyParser(&input); err != nil {
			return httpx.Error(c, fiber.StatusBadRequest, "Invalid request body")
		}

		community, err := svc.Update(user.ID, uint(id), input)
		if err != nil {
			return mapCommunityError(c, err)
		}
		return httpx.SuccessData(c, fiber.StatusOK, "Community updated successfully", "community", community)
	}
}

func Delete(svc *communities.Service) fiber.Handler {
	return func(c *fiber.Ctx) error {
		user, ok := httpx.CurrentUser(c)
		if !ok {
			return httpx.Error(c, fiber.StatusUnauthorized, "Unauthorized")
		}

		id, err := strconv.ParseUint(strings.TrimSpace(c.Params("id")), 10, 64)
		if err != nil || id == 0 {
			return httpx.Error(c, fiber.StatusBadRequest, "Invalid community id")
		}

		if err := svc.Delete(user.ID, uint(id)); err != nil {
			return mapCommunityError(c, err)
		}
		return httpx.Success(c, fiber.StatusOK, "Community deleted successfully", nil)
	}
}

func Join(svc *communities.Service) fiber.Handler {
	return func(c *fiber.Ctx) error {
		user, ok := httpx.CurrentUser(c)
		if !ok {
			return httpx.Error(c, fiber.StatusUnauthorized, "Unauthorized")
		}

		id, err := strconv.ParseUint(strings.TrimSpace(c.Params("id")), 10, 64)
		if err != nil || id == 0 {
			return httpx.Error(c, fiber.StatusBadRequest, "Invalid community id")
		}

		if err := svc.Join(user.ID, uint(id)); err != nil {
			return mapCommunityError(c, err)
		}
		return httpx.Success(c, fiber.StatusOK, "Joined community successfully", nil)
	}
}

func Leave(svc *communities.Service) fiber.Handler {
	return func(c *fiber.Ctx) error {
		user, ok := httpx.CurrentUser(c)
		if !ok {
			return httpx.Error(c, fiber.StatusUnauthorized, "Unauthorized")
		}

		id, err := strconv.ParseUint(strings.TrimSpace(c.Params("id")), 10, 64)
		if err != nil || id == 0 {
			return httpx.Error(c, fiber.StatusBadRequest, "Invalid community id")
		}

		if err := svc.Leave(user.ID, uint(id)); err != nil {
			return mapCommunityError(c, err)
		}
		return httpx.Success(c, fiber.StatusOK, "Left community successfully", nil)
	}
}

func Members(svc *communities.Service) fiber.Handler {
	return func(c *fiber.Ctx) error {
		id, err := strconv.ParseUint(strings.TrimSpace(c.Params("id")), 10, 64)
		if err != nil || id == 0 {
			return httpx.Error(c, fiber.StatusBadRequest, "Invalid community id")
		}

		page, perPage := parsePage(c)
		members, total, err := svc.Members(uint(id), page, perPage)
		if err != nil {
			return httpx.Error(c, fiber.StatusInternalServerError, "Failed to fetch members")
		}
		return httpx.Success(c, fiber.StatusOK, "Members fetched successfully", fiber.Map{
			"members": members,
			"total":   total,
		})
	}
}

func ListPending(svc *communities.Service) fiber.Handler {
	return func(c *fiber.Ctx) error {
		page, perPage := parsePage(c)
		result, err := svc.ListPending(page, perPage)
		if err != nil {
			return httpx.Error(c, fiber.StatusInternalServerError, "Failed to fetch pending communities")
		}
		return httpx.Success(c, fiber.StatusOK, "Pending communities fetched successfully", fiber.Map{
			"communities": result.Items,
			"total":       result.Total,
			"page":        result.Page,
			"pages":       result.Pages,
		})
	}
}

func Approve(svc *communities.Service) fiber.Handler {
	return func(c *fiber.Ctx) error {
		id, err := strconv.ParseUint(strings.TrimSpace(c.Params("id")), 10, 64)
		if err != nil || id == 0 {
			return httpx.Error(c, fiber.StatusBadRequest, "Invalid community id")
		}

		community, err := svc.SetStatus(uint(id), communities.StatusApproved)
		if err != nil {
			return mapCommunityError(c, err)
		}
		return httpx.SuccessData(c, fiber.StatusOK, "Community approved successfully", "community", community)
	}
}

func Reject(svc *communities.Service) fiber.Handler {
	return func(c *fiber.Ctx) error {
		id, err := strconv.ParseUint(strings.TrimSpace(c.Params("id")), 10, 64)
		if err != nil || id == 0 {
			return httpx.Error(c, fiber.StatusBadRequest, "Invalid community id")
		}

		community, err := svc.SetStatus(uint(id), communities.StatusRejected)
		if err != nil {
			return mapCommunityError(c, err)
		}
		return httpx.SuccessData(c, fiber.StatusOK, "Community rejected successfully", "community", community)
	}
}

func mapCommunityError(c *fiber.Ctx, err error) error {
	switch {
	case errors.Is(err, communities.ErrNotFound):
		return httpx.Error(c, fiber.StatusNotFound, "Community not found")
	case errors.Is(err, communities.ErrInvalidName):
		return httpx.ValidationError(c, "Name is required", fiber.Map{"name": "name is required"})
	case errors.Is(err, communities.ErrInvalidCategory):
		return httpx.ValidationError(c, "Invalid category type", fiber.Map{"category_type": "must be one of: movies, series, books, games, mixed"})
	case errors.Is(err, communities.ErrInvalidStatus):
		return httpx.Error(c, fiber.StatusBadRequest, "Invalid community status")
	case errors.Is(err, communities.ErrSlugTaken):
		return httpx.Error(c, fiber.StatusConflict, "A community with that slug already exists")
	case errors.Is(err, communities.ErrAlreadyMember):
		return httpx.Error(c, fiber.StatusConflict, "You are already a member of this community")
	case errors.Is(err, communities.ErrNotMember):
		return httpx.Error(c, fiber.StatusBadRequest, "You are not a member of this community")
	case errors.Is(err, communities.ErrBanned):
		return httpx.Error(c, fiber.StatusForbidden, "You are banned from this community")
	case errors.Is(err, communities.ErrRateLimitCommunity):
		return httpx.Error(c, fiber.StatusTooManyRequests, "You have created too many communities today")
	case errors.Is(err, communities.ErrForbidden):
		return httpx.Error(c, fiber.StatusForbidden, "You are not allowed to perform this action")
	default:
		return httpx.Error(c, fiber.StatusInternalServerError, "Something went wrong")
	}
}
