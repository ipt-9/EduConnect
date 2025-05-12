package main

import (
	"github.com/gorilla/mux"
	"github.com/ipt-9/EduConnect/DB"
	"github.com/ipt-9/EduConnect/Routes"
	"github.com/joho/godotenv"
	"github.com/rs/cors"
	"log"
	"net/http"
)

func main() {
	if err := DB.Connect(); err != nil {
		log.Fatalf("Fehler bei der Verbindung zur Datenbank: %v", err)
	}
	defer DB.Close()

	err := godotenv.Load("configuration.env")
	if err != nil {
		log.Fatal("❌ Fehler beim Laden der .env Datei")
	}

	r := mux.NewRouter()
	routes.InitJWT()
	handler := cors.AllowAll().Handler(r)
	r.HandleFunc("/register", routes.Register).Methods("POST", "OPTIONS")
	r.HandleFunc("/login", routes.Login).Methods("POST", "OPTIONS")
	r.HandleFunc("/protected", routes.Protected).Methods("GET", "OPTIONS")
	r.HandleFunc("/logout", routes.Logout).Methods("POST", "OPTIONS")
	r.HandleFunc("/verify-2fa", routes.Verify2FA).Methods("POST", "OPTIONS")

	r.HandleFunc("/users/me/submissions", routes.GetMySubmissionsHandler).Methods("GET", "OPTIONS")
	r.HandleFunc("/me", routes.Me).Methods("GET", "OPTIONS")
	r.HandleFunc("/my-courses", routes.GetMyCourses).Methods("GET", "OPTIONS")
	r.HandleFunc("/courses/{id}/tasks", routes.GetTasksByCourse).Methods("GET", "OPTIONS")
	r.HandleFunc("/submit", routes.SubmitTaskSolution).Methods("POST", "OPTIONS")
	r.HandleFunc("/tasks/{task_id}/submitted-code", routes.GetSubmittedCode).Methods("GET", "OPTIONS")
	r.HandleFunc("/tasks/{taskID}/tips", routes.GetUserTipsForTaskHandler).Methods("GET", "OPTIONS")

	r.HandleFunc("/groups", routes.GetUserGroupsHandler).Methods("GET", "OPTIONS")
	r.HandleFunc("/groups", routes.CreateGroupHandler).Methods("POST", "OPTIONS")
	r.HandleFunc("/groups/join", routes.JoinGroupByCodeHandler).Methods("POST", "OPTIONS")
	r.HandleFunc("/groups/{id}/members", routes.GetGroupMembersHandler).Methods("GET", "OPTIONS")
	r.HandleFunc("/groups/{groupID}/members/{userID}", routes.RemoveGroupMemberHandler).Methods("DELETE", "OPTIONS")
	r.HandleFunc("/groups/{groupID}/members/{userID}/role", routes.UpdateMemberRoleHandler).Methods("PUT", "OPTIONS")
	r.HandleFunc("/groups/{id}", routes.GetGroupByIDHandler).Methods("GET", "OPTIONS")
	r.HandleFunc("/groups/{groupID}/notifications", routes.GetGroupNotificationsHandler).Methods("GET", "OPTIONS")
	r.HandleFunc("/groups/{id}/share-submission", routes.ShareSubmissionHandler).Methods("POST", "OPTIONS")
	r.HandleFunc("/last-course", routes.GetLastVisitedCourseHandler).Methods("GET", "OPTIONS")
	r.HandleFunc("/dashboard-overview", routes.GetDashboardOverviewHandler).Methods("GET", "OPTIONS")
	r.HandleFunc("/progress/overview", routes.GetUserProgressOverview)

	r.HandleFunc("/groups/{groupID}/messages", routes.GetGroupMessagesHandler).Methods("GET", "OPTIONS")
	r.HandleFunc("/ws/groups/{groupID}/chat", routes.HandleGroupChatWS)

	//Payment routes
	r.HandleFunc("/create-checkout-session", routes.CreateCheckoutSession).Methods("POST", "OPTIONS")
	r.HandleFunc("/session-status", routes.RetrieveCheckoutSession).Methods("GET", "OPTIONS")

	log.Println("🚀 Server läuft auf http://localhost:8080")
	http.ListenAndServe(":8080", handler)
}
