package commentscontroller

import (
	"errors"
	"strconv"
	"strings"

	"github.com/gofiber/fiber/v2"

	comments "trendflix/services/comments"
	posts "trendflix/services/posts"
	"trendflix/utils/httpx"
)

func List(svc *comments.Service, postSvc *posts.Service) fiber.Handler {
	return func(c *fiber.Ctx) error {
		postID, err := strconv.ParseUint(strings.TrimSpace(c.Params("id")), 10, 64)
		if err != nil || postID == 0 {
			return httpx.Error(c, fiber.StatusBadRequest, "Invalid post id")
		}

		post, err := postSvc.GetByID(uint(postID))
		if err != nil {
			if errors.Is(err, posts.ErrNotFound) {
				return httpx.Error(c, fiber.StatusNotFound, "Post not found")
			}
			return httpx.Error(c, fiber.StatusInternalServerError, "Failed to fetch post")
		}
		_ = post

		tree, err := svc.Tree(uint(postID))
		if err != nil {
			return httpx.Error(c, fiber.StatusInternalServerError, "Failed to fetch comments")
		}
		return httpx.Success(c, fiber.StatusOK, "Comments fetched successfully", fiber.Map{
			"comments": tree,
		})
	}
}

func Create(svc *comments.Service, postSvc *posts.Service) fiber.Handler {
	return func(c *fiber.Ctx) error {
		user, ok := httpx.CurrentUser(c)
		if !ok {
			return httpx.Error(c, fiber.StatusUnauthorized, "Unauthorized")
		}

		postID, err := strconv.ParseUint(strings.TrimSpace(c.Params("id")), 10, 64)
		if err != nil || postID == 0 {
			return httpx.Error(c, fiber.StatusBadRequest, "Invalid post id")
		}

		var input comments.CreateInput
		if err := c.BodyParser(&input); err != nil {
			return httpx.Error(c, fiber.StatusBadRequest, "Invalid request body")
		}

		post, err := postSvc.GetByID(uint(postID))
		if err != nil {
			if errors.Is(err, posts.ErrNotFound) {
				return httpx.Error(c, fiber.StatusNotFound, "Post not found")
			}
			return httpx.Error(c, fiber.StatusInternalServerError, "Failed to fetch post")
		}

		if err := postSvc.EnsureCanComment(post, user.ID); err != nil {
			return mapCommentError(c, err)
		}

		comment, err := svc.Create(user.ID, uint(postID), input)
		if err != nil {
			return mapCommentError(c, err)
		}
		return httpx.SuccessData(c, fiber.StatusCreated, "Comment added successfully", "comment", comment)
	}
}

func Update(svc *comments.Service) fiber.Handler {
	return func(c *fiber.Ctx) error {
		user, ok := httpx.CurrentUser(c)
		if !ok {
			return httpx.Error(c, fiber.StatusUnauthorized, "Unauthorized")
		}

		id, err := strconv.ParseUint(strings.TrimSpace(c.Params("id")), 10, 64)
		if err != nil || id == 0 {
			return httpx.Error(c, fiber.StatusBadRequest, "Invalid comment id")
		}

		var input comments.UpdateInput
		if err := c.BodyParser(&input); err != nil {
			return httpx.Error(c, fiber.StatusBadRequest, "Invalid request body")
		}

		comment, err := svc.Update(user.ID, uint(id), input)
		if err != nil {
			return mapCommentError(c, err)
		}
		return httpx.SuccessData(c, fiber.StatusOK, "Comment updated successfully", "comment", comment)
	}
}

func Delete(svc *comments.Service) fiber.Handler {
	return func(c *fiber.Ctx) error {
		user, ok := httpx.CurrentUser(c)
		if !ok {
			return httpx.Error(c, fiber.StatusUnauthorized, "Unauthorized")
		}

		id, err := strconv.ParseUint(strings.TrimSpace(c.Params("id")), 10, 64)
		if err != nil || id == 0 {
			return httpx.Error(c, fiber.StatusBadRequest, "Invalid comment id")
		}

		if err := svc.Delete(user.ID, uint(id)); err != nil {
			return mapCommentError(c, err)
		}
		return httpx.Success(c, fiber.StatusOK, "Comment deleted successfully", nil)
	}
}

func mapCommentError(c *fiber.Ctx, err error) error {
	switch {
	case errors.Is(err, comments.ErrNotFound):
		return httpx.Error(c, fiber.StatusNotFound, "Comment not found")
	case errors.Is(err, comments.ErrInvalidBody):
		return httpx.ValidationError(c, "Comment body is required", fiber.Map{"body": "body is required"})
	case errors.Is(err, comments.ErrParentNotFound):
		return httpx.Error(c, fiber.StatusBadRequest, "Parent comment not found")
	case errors.Is(err, comments.ErrForbidden):
		return httpx.Error(c, fiber.StatusForbidden, "You are not allowed to perform this action")
	case errors.Is(err, comments.ErrRateLimitComment):
		return httpx.Error(c, fiber.StatusTooManyRequests, "You are commenting too fast, try again later")
	case errors.Is(err, posts.ErrLocked):
		return httpx.Error(c, fiber.StatusForbidden, "This post is locked")
	case errors.Is(err, posts.ErrMustJoin):
		return httpx.Error(c, fiber.StatusForbidden, "You must join this community before commenting")
	case errors.Is(err, posts.ErrBanned):
		return httpx.Error(c, fiber.StatusForbidden, "You are banned from this community")
	default:
		return httpx.Error(c, fiber.StatusInternalServerError, "Something went wrong")
	}
}
