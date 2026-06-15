package routers

import (
	"github.com/gofiber/fiber/v2"

	commentscontroller "trendflix/controller/comments_controller"
	communitiescontroller "trendflix/controller/communities_controller"
	postscontroller "trendflix/controller/posts_controller"
	votescontroller "trendflix/controller/votes_controller"
	"trendflix/middleware"
	communities "trendflix/services/communities"
	comments "trendflix/services/comments"
	posts "trendflix/services/posts"
	votes "trendflix/services/votes"
)

func RegisterCommunityRoutes(app *fiber.App) {
	communitySvc := communities.NewService()
	postSvc := posts.NewService(communitySvc)
	commentSvc := comments.NewService()
	voteSvc := votes.NewService()

	api := app.Group("/api")

	// Public read endpoints
	pub := api.Group("/communities")
	pub.Get("", communitiescontroller.List(communitySvc))
	pub.Get("/popular", communitiescontroller.Popular(communitySvc))
	pub.Get("/recommended", communitiescontroller.Recommended(communitySvc))
	pub.Get("/:slug", communitiescontroller.GetBySlug(communitySvc))
	pub.Get("/:id/posts", postscontroller.ListByCommunity(postSvc))
	pub.Get("/:id/members", communitiescontroller.Members(communitySvc))

	// Authenticated community management
	authCom := api.Group("/communities", middleware.Authenticate)
	authCom.Post("", communitiescontroller.Create(communitySvc))
	authCom.Put("/:id", communitiescontroller.Update(communitySvc))
	authCom.Delete("/:id", communitiescontroller.Delete(communitySvc))
	authCom.Post("/:id/join", communitiescontroller.Join(communitySvc))
	authCom.Post("/:id/leave", communitiescontroller.Leave(communitySvc))

	// Posts (create under a community requires auth)
	api.Post("/communities/:id/posts", middleware.Authenticate, postscontroller.Create(postSvc))

	// Posts read / manage
	api.Get("/posts/:id", postscontroller.GetByID(postSvc))
	authPosts := api.Group("/posts", middleware.Authenticate)
	authPosts.Put("/:id", postscontroller.Update(postSvc))
	authPosts.Delete("/:id", postscontroller.Delete(postSvc))

	// Comments
	api.Get("/posts/:id/comments", commentscontroller.List(commentSvc, postSvc))
	api.Post("/posts/:id/comments", middleware.Authenticate, commentscontroller.Create(commentSvc, postSvc))
	authComments := api.Group("/comments", middleware.Authenticate)
	authComments.Put("/:id", commentscontroller.Update(commentSvc))
	authComments.Delete("/:id", commentscontroller.Delete(commentSvc))

	// Votes
	api.Post("/posts/:id/vote", middleware.Authenticate, votescontroller.VotePost(voteSvc))
	api.Post("/comments/:id/vote", middleware.Authenticate, votescontroller.VoteComment(voteSvc))

	// Item discussions
	api.Get("/items/:type/:id/discussions", postscontroller.ListByItem(postSvc))
}
