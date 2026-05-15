package main

import (
	"context"
	"log"
	"net"
	"os"
	"strings"

	"github.com/HLLC-MFU/hllc-workshop-backend/auth"
	"github.com/HLLC-MFU/hllc-workshop-backend/course"
	"github.com/HLLC-MFU/hllc-workshop-backend/database"
	"github.com/HLLC-MFU/hllc-workshop-backend/item"
	"github.com/HLLC-MFU/hllc-workshop-backend/major"
	"github.com/HLLC-MFU/hllc-workshop-backend/push"
	"github.com/HLLC-MFU/hllc-workshop-backend/ws"
	fws "github.com/fasthttp/websocket"
	"github.com/gofiber/fiber/v3"
	"github.com/gofiber/fiber/v3/middleware/cors"
	"github.com/joho/godotenv"
	"github.com/valyala/fasthttp"
)

func main() {
	_ = godotenv.Load()

	db, err := database.Connect(context.Background())
	if err != nil {
		log.Fatal("mongo connect:", err)
	}

	app := fiber.New(fiber.Config{
		BodyLimit: 10 * 1024 * 1024,
	})

	// CORS — เปิดให้ FE ที่ port 3001 เรียกได้
	app.Use(cors.New(cors.Config{
		AllowOrigins: getAllowedOrigins(),
		AllowMethods: []string{"GET", "POST", "PUT", "DELETE"},
		AllowHeaders: []string{"Content-Type", "Authorization"},
	}))

	app.Get("/health", func(c fiber.Ctx) error {
		return c.JSON(fiber.Map{"ok": true})
	})

	// Server config for QR code generation (LAN IP + guest URL)
	app.Get("/config", func(c fiber.Ctx) error {
		return c.JSON(fiber.Map{
			"guestUrl": getGuestURL(c),
		})
	})

	// wire major: repository <- service <- handler
	majorRepo := major.NewRepository(db.Collection("majors"))
	majorSvc := major.NewService(majorRepo)
	majorH := major.NewHandler(majorSvc)
	majorH.RegisterRoutes(app)

	// wire course (เช้าทำ major, บ่ายต่อยอดเป็น course)
	courseRepo := course.NewRepository(db.Collection("courses"))
	courseSvc := course.NewService(courseRepo)
	courseH := course.NewHandler(courseSvc)
	courseH.RegisterRoutes(app)

	authRepo := auth.NewRepository(db.Collection("auths"))
	authSvc := auth.NewService(authRepo)
	authH := auth.NewHandler(authSvc)
	authH.RegisterRoutes(app)

	itemRepo := item.NewRepository(db.Collection("display_items"))
	itemSvc := item.NewService(itemRepo)
	itemH := item.NewHandler(itemSvc)
	itemH.RegisterRoutes(app)

	pushRrpo := push.NewRepository(db.Collection("pushes"))
	pushSvc := push.NewService(pushRrpo)
	pushH := push.NewHandler(pushSvc)
	pushH.RegisterRoutes(app)

	// websocket

	port := os.Getenv("PORT")
	if port == "" {
		port = "3000"
	}
	upgrader := fws.FastHTTPUpgrader{
		CheckOrigin: func(ctx *fasthttp.RequestCtx) bool {
			return true
		},
	}
	app.Get("/ws", func(c fiber.Ctx) error {
		if !c.IsWebSocket() {
			return fiber.ErrUpgradeRequired
		}

		return upgrader.Upgrade(c.RequestCtx(), func(conn *fws.Conn) {
			ws.Mutex.Lock()
			ws.Clients[conn] = true
			ws.Mutex.Unlock()

			defer func() {
				ws.Mutex.Lock()
				delete(ws.Clients, conn)
				ws.Mutex.Unlock()

				conn.Close()
			}()

			for {
				if _, _, err := conn.ReadMessage(); err != nil {
					break
				}
			}
		})
	})
	log.Println("listening on :" + port)
	log.Fatal(app.Listen(":" + port))
}

// getLANIP returns the first non-loopback IPv4 address.
// Falls back to empty string if no suitable interface is found.
func getLANIP() string {
	interfaces, err := net.Interfaces()
	if err != nil {
		return ""
	}
	fallback := ""
	for _, iface := range interfaces {
		if iface.Flags&net.FlagUp == 0 || iface.Flags&net.FlagLoopback != 0 {
			continue
		}
		addrs, err := iface.Addrs()
		if err != nil {
			continue
		}
		for _, addr := range addrs {
			ip := ipv4FromAddr(addr)
			if ip == nil || ip.IsLoopback() || ip.IsLinkLocalUnicast() {
				continue
			}
			if ip.IsPrivate() {
				return ip.String()
			}
			if fallback == "" {
				fallback = ip.String()
			}
		}
	}
	return fallback
}

func ipv4FromAddr(addr net.Addr) net.IP {
	switch value := addr.(type) {
	case *net.IPNet:
		return value.IP.To4()
	case *net.IPAddr:
		return value.IP.To4()
	default:
		return nil
	}
}

func getAllowedOrigins() []string {
	frontendPort := getFrontendPort()
	origins := []string{
		"http://localhost:" + frontendPort,
		"http://127.0.0.1:" + frontendPort,
	}
	if lanIP := getLANIP(); lanIP != "" {
		origins = append(origins, "http://"+lanIP+":"+frontendPort)
	}

	origins = append(origins, splitOrigins(os.Getenv("ALLOWED_ORIGIN"))...)
	origins = append(origins, splitOrigins(os.Getenv("ALLOWED_ORIGINS"))...)
	return uniqueNonEmpty(origins)
}

func getGuestURL(c fiber.Ctx) string {
	if publicURL := strings.TrimRight(strings.TrimSpace(os.Getenv("PUBLIC_FRONTEND_URL")), "/"); publicURL != "" {
		return publicURL + "/guest"
	}
	if frontendURL := strings.TrimRight(strings.TrimSpace(os.Getenv("FRONTEND_ORIGIN")), "/"); frontendURL != "" {
		return frontendURL + "/guest"
	}
	if lanIP := getLANIP(); lanIP != "" {
		return "http://" + lanIP + ":" + getFrontendPort() + "/guest"
	}
	if origin := strings.TrimRight(strings.TrimSpace(c.Get("Origin")), "/"); origin != "" {
		return origin + "/guest"
	}
	return "http://localhost:" + getFrontendPort() + "/guest"
}

func getFrontendPort() string {
	port := strings.TrimSpace(os.Getenv("FRONTEND_PORT"))
	if port == "" {
		return "3001"
	}
	return port
}

func splitOrigins(raw string) []string {
	parts := strings.Split(raw, ",")
	origins := make([]string, 0, len(parts))
	for _, part := range parts {
		origin := strings.TrimRight(strings.TrimSpace(part), "/")
		if origin != "" {
			origins = append(origins, origin)
		}
	}
	return origins
}

func uniqueNonEmpty(values []string) []string {
	seen := make(map[string]bool, len(values))
	out := make([]string, 0, len(values))
	for _, value := range values {
		value = strings.TrimSpace(value)
		if value == "" || seen[value] {
			continue
		}
		seen[value] = true
		out = append(out, value)
	}
	return out
}
