package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os"

	"github.com/coreos/go-oidc/v3/oidc"
)

func main() {
	if len(os.Args) < 2 {
		log.Fatal("usage: go run main.go <access_token>")
	}
	tokenString := os.Args[1]

	ctx := context.Background()

	// Phải khớp với issuer trong claim "iss" của token.
	// Với realm "demo" trên Keycloak chạy ở localhost:8080:
	issuer := "http://localhost:8090/realms/multi-agent"

	// NewProvider sẽ tự fetch /.well-known/openid-configuration
	// và từ đó biết jwks_uri để lấy public key.
	provider, err := oidc.NewProvider(ctx, issuer)
	if err != nil {
		log.Fatalf("không tạo được provider (kiểm tra Keycloak đã chạy chưa, issuer URL đúng chưa): %v", err)
	}

	// SkipClientIDCheck = true vì access token mặc định của Keycloak
	// có "aud": "account", không phải client_id bạn dùng để lấy token.
	// Trong production nên cấu hình audience mapper rồi bỏ dòng này.
	verifier := provider.Verifier(&oidc.Config{
		SkipClientIDCheck: true,
	})

	idToken, err := verifier.Verify(ctx, tokenString)
	if err != nil {
		log.Fatalf("token không hợp lệ: %v", err)
	}

	var claims map[string]interface{}
	if err := idToken.Claims(&claims); err != nil {
		log.Fatalf("không parse được claims: %v", err)
	}

	fmt.Println("Token hợp lệ. Claims:")
	pretty, _ := json.MarshalIndent(claims, "", "  ")
	fmt.Println(string(pretty))
}
