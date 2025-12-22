package server

import (
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/cors"
	adminauth "github.com/varnit-ta/PlacementLog/internal/adminAuth"
	"github.com/varnit-ta/PlacementLog/internal/db"
	placements "github.com/varnit-ta/PlacementLog/internal/placements"
	"github.com/varnit-ta/PlacementLog/internal/posts"
	userauth "github.com/varnit-ta/PlacementLog/internal/userAuth"
	"github.com/varnit-ta/PlacementLog/pkg/middleware"
)

type App struct {
	userAuthHandler   *userauth.UserAuthHandler
	postHandler       *posts.PostsHandler
	adminHandler      *adminauth.AdminAuthHandler
	placementsHandler *placements.PlacementsHandler
}

func InitApp() (*App, error) {
	db, err := db.InitDatabse()

	if err != nil {
		return nil, err
	}

	userAuthRepo := userauth.NewUserAuthRepo(db)
	userAuthService := userauth.NewUserAuthService(userAuthRepo)
	userAuthHandler := userauth.NewUserAuthHandler(userAuthService)

	postRepo := posts.NewPostsRepo(db)
	postService := posts.NewPostsService(postRepo)
	postHandler := posts.NewPostsHandler(postService)

	adminRepo := adminauth.NewAdminRepo(db)
	adminService := adminauth.NewAdminService(adminRepo)
	adminHandler := adminauth.NewAdminAuthHandler(adminService)

	placementsRepo := placements.NewPlacementsRepo(db)
	placementsService := placements.NewPlacementsService(placementsRepo)
	placementsHandler := placements.NewPlacementsHandler(placementsService)

	return &App{
		userAuthHandler:   userAuthHandler,
		postHandler:       postHandler,
		adminHandler:      adminHandler,
		placementsHandler: placementsHandler,
	}, nil
}

func (a App) Routes() http.Handler {
	r := chi.NewRouter()

	r.Use(cors.Handler(cors.Options{
		AllowedOrigins:   []string{"https://*", "http://*"},
		AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowedHeaders:   []string{"Accept", "Authorization", "Content-Type", "X-CSRF-Token", "X-User-ID"},
		ExposedHeaders:   []string{"Link"},
		AllowCredentials: true,
		MaxAge:           300,
	}))

	// Public routes (no authentication required)
	r.Group(func(r chi.Router) {
		r.Post("/auth/login", a.userAuthHandler.Login)
		r.Post("/auth/register", a.userAuthHandler.Register)
		r.Post("/admin/login", a.adminHandler.Login)
		r.Get("/placements", a.placementsHandler.GetAllPlacements)
		r.Get("/placements/company-branch", a.placementsHandler.GetCompanyBranchMap)
		r.Get("/placements/branch-company", a.placementsHandler.GetBranchCompanyMap)
		r.Get("/posts", a.postHandler.GetAll)
	})

	// User authenticated routes
	r.Group(func(r chi.Router) {
		r.Use(middleware.UserAuthMiddleware)

		r.Post("/posts", a.postHandler.AddPost)
		r.Put("/posts", a.postHandler.UpdatePost)
		r.Delete("/posts", a.postHandler.DeletePost)
		r.Get("/posts/user", a.postHandler.GetByUser)
	})

	// Admin authenticated routes
	r.Group(func(r chi.Router) {
		r.Use(middleware.AdminAuthMiddleware)

		r.Post("/admin/register", a.adminHandler.Register)
		r.Get("/admin/posts", a.postHandler.GetAllPostsForAdmin)
		r.Put("/admin/posts/review", a.postHandler.ReviewPost)
		r.Delete("/admin/posts", a.postHandler.DeletePostAsAdmin)
		r.Post("/admin/placements", a.placementsHandler.AddPlacement)
	})

	return r
}
