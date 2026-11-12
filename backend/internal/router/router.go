package router

import (
	"net/http"
	"time"

	"github.com/MaryJane-09/noxorbit/backend/config"
	"github.com/MaryJane-09/noxorbit/backend/internal/auth"
	"github.com/MaryJane-09/noxorbit/backend/internal/email"
	"github.com/MaryJane-09/noxorbit/backend/internal/health"
	"github.com/MaryJane-09/noxorbit/backend/internal/membership"
	"github.com/MaryJane-09/noxorbit/backend/internal/otp"
	"github.com/MaryJane-09/noxorbit/backend/internal/project"
	"github.com/MaryJane-09/noxorbit/backend/internal/ratelimit"
	"github.com/MaryJane-09/noxorbit/backend/internal/register"
	"github.com/MaryJane-09/noxorbit/backend/internal/session"
	"github.com/MaryJane-09/noxorbit/backend/internal/task"
	"github.com/MaryJane-09/noxorbit/backend/internal/user"
	"github.com/MaryJane-09/noxorbit/backend/internal/workspace"
	"github.com/jackc/pgx/v5/pgxpool"
)

func corsMiddleware(next http.Handler) http.Handler {
	// Vite normally starts on 5173 and may switch to 5174 if 5173 is in use.
	// Only trusted local development origins may make credentialed requests.
	allowedOrigins := map[string]bool{
		"http://localhost:5173": true,
		"http://localhost:5174": true,
		"http://127.0.0.1:5173": true,
		"http://127.0.0.1:5174": true,
	}

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Add("Vary", "Origin")
		origin := r.Header.Get("Origin")
		if allowedOrigins[origin] {
			w.Header().Set("Access-Control-Allow-Origin", origin)
			w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, PATCH, DELETE, OPTIONS")
			w.Header().Set("Access-Control-Allow-Headers", "Content-Type")
			w.Header().Set("Access-Control-Allow-Credentials", "true")
		}

		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusNoContent)
			return
		}

		next.ServeHTTP(w, r)
	})
}

func New(pool *pgxpool.Pool) http.Handler {

	sender := email.EmailSender{
		GmailAddress:     config.AppConfig.GmailAddress,
		GmailAppPassword: config.AppConfig.GmailAppPassword,
		SMTPHost:         "smtp.gmail.com",
		SMTPPort:         "587",
	}
	repo := user.NewRepository(pool)
	otpRepo := otp.NewRepository(pool)
	sessionRepo := session.NewRepository(pool)
	limits := ratelimit.New(pool)
	workspaceHandler := workspace.NewHandler(workspace.NewRepository(pool))
	memberHandler := membership.NewHandler(membership.NewRepository(pool))
	invitationStore := membership.NewLinkRepository(pool)
	invitationHandler := membership.NewInvitationHandler(invitationStore)
	managementHandler := membership.NewManagementHandler(invitationStore,membership.NewMemberManagementRepository(pool))
	projectHandler := project.NewHandler(project.NewRepository(pool))
	taskHandler := task.NewHandler(task.NewRepository(pool))
	mux := http.NewServeMux()
	mux.HandleFunc("/health", health.HealthHandler)
	// Fixed 15-minute IP quotas limit request floods before expensive work.
	// These use the TCP peer, not untrusted forwarded headers.
	const window = 15 * time.Minute
	mux.Handle("/register", limits.LimitIP("register", 12, window, register.RegisterHandler(otpRepo, &sender)))
	mux.HandleFunc("/me", auth.RequireAuth(sessionRepo, auth.MeHandler(repo)))
	// Only authenticated users can create or list their own workspaces.
	// The repository enforces the per-user membership filter.
	mux.HandleFunc("/api/v1/workspaces", auth.RequireAuth(sessionRepo, workspaceHandler.Collection))
	mux.HandleFunc("/api/v1/workspaces/{id}", auth.RequireAuth(sessionRepo, workspaceHandler.Detail))
	mux.HandleFunc("/api/v1/workspaces/{workspace_id}/members", auth.RequireAuth(sessionRepo, memberHandler.List))
	mux.HandleFunc("/api/v1/workspaces/{workspace_id}/invitations", auth.RequireAuth(sessionRepo, invitationHandler.Create))
	mux.HandleFunc("/api/v1/invitations/accept", auth.RequireAuth(sessionRepo, invitationHandler.Accept))
	mux.HandleFunc("/api/v1/workspaces/{workspace_id}/invitations/manage", auth.RequireAuth(sessionRepo, managementHandler.List))
	mux.HandleFunc("/api/v1/workspaces/{workspace_id}/invitations/{invitation_id}", auth.RequireAuth(sessionRepo, managementHandler.Revoke))
	mux.HandleFunc("/api/v1/workspaces/{workspace_id}/members/{member_id}", auth.RequireAuth(sessionRepo, managementHandler.Remove))
	// All project routes are workspace-scoped; repositories enforce membership
	// or owner permissions on the data read or mutation itself.
	mux.HandleFunc("/api/v1/workspaces/{workspace_id}/projects", auth.RequireAuth(sessionRepo, projectHandler.Collection))
	mux.HandleFunc("/api/v1/workspaces/{workspace_id}/projects/{project_id}", auth.RequireAuth(sessionRepo, projectHandler.Detail))
	mux.HandleFunc("/api/v1/workspaces/{workspace_id}/projects/{project_id}/status", auth.RequireAuth(sessionRepo, projectHandler.Status))
	// All task endpoints require a session; repository writes enforce the task role rules.
	mux.HandleFunc("/api/v1/workspaces/{workspace_id}/projects/{project_id}/tasks", auth.RequireAuth(sessionRepo, taskHandler.Collection))
	mux.HandleFunc("/api/v1/workspaces/{workspace_id}/projects/{project_id}/tasks/{task_id}", auth.RequireAuth(sessionRepo, taskHandler.Detail))
	mux.HandleFunc("/api/v1/workspaces/{workspace_id}/projects/{project_id}/tasks/{task_id}/status", auth.RequireAuth(sessionRepo, taskHandler.Status))
	mux.HandleFunc("/logout", auth.LogoutHandler(sessionRepo))
	mux.Handle("/verify-otp", limits.LimitIP("verify-otp", 40, window, register.VerifyOTPHandler(repo, otpRepo)))
	mux.Handle("/login", limits.LimitIP("login", 60, window, auth.LoginHandler(repo, sessionRepo, limits)))
	mux.Handle("/forgot-password", limits.LimitIP("forgot-password", 12, window, auth.ForgotPasswordHandler(repo, otpRepo, &sender)))
	mux.Handle("/reset-password", limits.LimitIP("reset-password", 40, window, auth.ResetPasswordHandler(otpRepo)))
	return corsMiddleware(mux)
}
