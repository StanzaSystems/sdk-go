package fiberstanza

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/StanzaSystems/sdk-go/logging"
	"github.com/StanzaSystems/sdk-go/stanza"

	"github.com/gofiber/fiber/v2"
	"github.com/valyala/fasthttp/fasthttpadaptor"
)

// Client defines options for a new Stanza Client
type Client struct {
	// Required
	APIKey string // customer generated API key

	// Optional
	Name        string // defines applications name
	Release     string // defines applications version
	Environment string // defines applications environment
	StanzaHub   string // host:port (ipv4, ipv6, or resolvable hostname)
	DataSource  string // local:<path>, consul:<key>, or grpc:host:port
}

// Decorator defines the config for fiberstanza middleware.
type Decorator struct {
	Name string // optional (but required if you want to use multiple Decorators)
}

// New creates a new fiberstanza middleware fiber.Handler
func New(d Decorator) fiber.Handler {
	h, err := stanza.NewHttpInboundHandler(d.Name)
	if err != nil {
		logging.Error(fmt.Errorf("failed to initialize new http inbound meters: %v", err))
	}

	return func(c *fiber.Ctx) error {
		start := time.Now()
		savedCtx, cancel := context.WithCancel(c.UserContext())

		h.Meter().ActiveRequests.Add(savedCtx, 1, h.Meter().Attributes...)
		defer func() {
			h.Meter().Duration.Record(savedCtx, float64(time.Since(start).Microseconds())/1000, h.Meter().Attributes...)
			h.Meter().RequestSize.Record(savedCtx, int64(len(c.Request().Body())), h.Meter().Attributes...)
			h.Meter().ResponseSize.Record(savedCtx, int64(len(c.Response().Body())), h.Meter().Attributes...)
			h.Meter().ActiveRequests.Add(savedCtx, -1, h.Meter().Attributes...)
			c.SetUserContext(savedCtx)
			cancel()
		}()

		// TODO(msg): implement HttpInboundHandler as fasthttp handler instead of converting to net/http?
		var req http.Request
		if err := fasthttpadaptor.ConvertRequest(c.Context(), &req, true); err != nil {
			logging.Error(fmt.Errorf("failed to convert request from fasthttp: %v", err))
			return c.Next() // log error and fail open
		}
		ctx, status := h.VerifyServingCapacity(&req, c.Route().Path)
		if status != http.StatusOK {
			return c.SendStatus(status)
		}
		c.SetUserContext(ctx)
		return c.Next()
	}
}

func Init(ctx context.Context, client Client) (func(), error) {
	return stanza.Init(ctx, stanza.ClientOptions(client))
}
