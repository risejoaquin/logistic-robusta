package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"strings"
)

func basicAuth(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Bypass CORS preflight requests
		if r.Method == "OPTIONS" {
			next(w, r)
			return
		}

		user := strings.TrimSpace(strings.Trim(os.Getenv("ADMIN_USER"), "\""))
		pass := strings.TrimSpace(strings.Trim(os.Getenv("ADMIN_PASS"), "\""))

		u, p, ok := r.BasicAuth()
		if !ok || u != user || p != pass {
			w.Header().Set("WWW-Authenticate", `Basic realm="Restricted Area"`)
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}
		next(w, r)
	}
}

func main() {
	adminUser := strings.TrimSpace(strings.Trim(os.Getenv("ADMIN_USER"), "\""))
	adminPass := strings.TrimSpace(strings.Trim(os.Getenv("ADMIN_PASS"), "\""))
	if adminUser == "" || adminPass == "" {
		log.Fatal("CRITICAL ERROR: Variables de entorno ADMIN_USER y ADMIN_PASS son obligatorias (Fail-Fast).")
	}

	databaseURL := strings.TrimSpace(strings.Trim(os.Getenv("DATABASE_URL"), "\""))
	if databaseURL != "" {
		InitDB(databaseURL)
	} else {
		log.Println("WARNING: DATABASE_URL no provista")
	}

	if DB != nil {
		if err := DB.Ping(context.Background()); err == nil {
			log.Println("✅ Verificación de DB: Conexión y Ping exitosos antes de iniciar el servidor.")
		} else {
			log.Printf("⚠️ Verificación de DB: El Ping a la base de datos falló: %v\n", err)
		}
	} else {
		log.Println("⚠️ Verificación de DB: Operando sin conexión a base de datos activa (DB es nil).")
	}

	port := os.Getenv("PORT")
	if port == "" {
		port = "3000"
	}

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/" {
			http.Redirect(w, r, "/admin", http.StatusFound)
			return
		}
		http.NotFound(w, r)
	})
	
	// Webhook remains unprotected (public for Meta)
	http.HandleFunc("/webhook", WebhookHandler)
	
	// Protected UI Routes
	http.HandleFunc("/monitor", basicAuth(HandleMonitorInterface))
	http.HandleFunc("/monitor/stream", basicAuth(HandleMonitorStream))
	http.HandleFunc("/admin", basicAuth(HandleAdminInterface))

	// CORS preflight global para /api/ en adelante
	http.HandleFunc("OPTIONS /api/", func(w http.ResponseWriter, r *http.Request) {
		if enableCors(&w, r) {
			w.WriteHeader(http.StatusOK)
		} else {
			http.Error(w, "Forbidden", http.StatusForbidden)
		}
	})

	// Protected API Routes
	http.HandleFunc("GET /api/test_error", HandleTestGeminiError)
	http.HandleFunc("GET /api/orders", basicAuth(GetOrdersAPI))
	http.HandleFunc("POST /api/orders", basicAuth(CreateOrderAPI))
	http.HandleFunc("PUT /api/orders", basicAuth(UpdateOrderAPI))
	http.HandleFunc("DELETE /api/orders", basicAuth(DeleteOrderAPI))
	
	http.HandleFunc("GET /api/inventory", basicAuth(GetInventoryAPI))
	http.HandleFunc("POST /api/inventory", basicAuth(CreateInventoryAPI))
	http.HandleFunc("PUT /api/inventory", basicAuth(UpdateInventoryAPI))
	http.HandleFunc("DELETE /api/inventory", basicAuth(DeleteInventoryAPI))
	
	http.HandleFunc("GET /api/dashboard", basicAuth(GetDashboardAPI))
	http.HandleFunc("GET /api/system_status", basicAuth(GetSystemStatusAPI))
	
	http.HandleFunc("GET /api/tickets", basicAuth(GetTicketsAPI))
	http.HandleFunc("PUT /api/tickets", basicAuth(UpdateTicketAPI))
	
	http.HandleFunc("GET /api/accounting", basicAuth(GetAccountingAPI))
	http.HandleFunc("POST /api/accounting", basicAuth(CreateAccountingAPI))
	
	http.HandleFunc("GET /api/menu", basicAuth(GetMenuAPI))
	http.HandleFunc("POST /api/menu", basicAuth(CreateMenuAPI))
	http.HandleFunc("PUT /api/menu", basicAuth(UpdateMenuAPI))
	http.HandleFunc("DELETE /api/menu", basicAuth(DeleteMenuAPI))

	log.Println("⚡ Webhook escuchando en: http://localhost:" + port + "/webhook")
	log.Println("🖥️ Monitor disponible en: http://localhost:" + port + "/monitor")
	log.Println("📊 Panel Admin disponible en: http://localhost:" + port + "/admin")
	
	err := http.ListenAndServe("0.0.0.0:"+port, nil)
	if err != nil {
		log.Fatalf("CRITICAL ERROR: El servidor colapsó: %v\n", err)
	}
}
