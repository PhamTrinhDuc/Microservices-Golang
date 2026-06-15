sequenceDiagram
    actor User

    participant MA as Multi-Agent
    participant AUTH as Auth Service (Internal)
    participant MCP1 as MCP Server (Internal)
    participant BE as Backend API

    participant MCP2 as MCP Server (3rd Party)
    participant AS as Authorization Server (3rd Party)

    %% ==========================================================
    %% INTERNAL FLOW
    %% ==========================================================

    rect rgb(240,248,255)
        Note over AUTH,BE: Internal ecosystem - Shared issuer

        User->>MA: User request

        MA->>AUTH: Login / Exchange credential
        AUTH-->>MA: Internal JWT\n(iss=auth.internal)

        MA->>MCP1: Tool Call + Internal JWT\n(aud=mcp-internal)

        MCP1->>MCP1: Verify Signature
        MCP1->>MCP1: Validate iss
        MCP1->>MCP1: Validate aud

        MCP1->>BE: API Call + Internal JWT\n(aud=backend)

        BE->>BE: Verify Signature
        BE->>BE: Validate iss
        BE->>BE: Validate aud

        BE-->>MCP1: Business Response

        MCP1-->>MA: Tool Result
        MA-->>User: Final Answer
    end

    %% ==========================================================
    %% THIRD PARTY MCP FLOW
    %% ==========================================================

    rect rgb(255,250,240)
        Note over MA,AS: First call to third-party MCP (no token yet)

        MA->>MCP2: Tool Call

        MCP2-->>MA: 401 Unauthorized\nWWW-Authenticate Metadata

        MA->>MCP2: Discover OAuth Metadata

        MCP2-->>MA: authorization_endpoint\ntoken_endpoint\nscopes

        Note over MA: Generate PKCE\ncode_verifier\ncode_challenge\nstate

        MA->>AS: Authorization Request\n(client_id, scope,\ncode_challenge, state)

        AS-->>User: Login Page

        User->>AS: Login + Consent

        AS-->>MA: Authorization Code

        MA->>AS: Token Exchange\n(code + code_verifier)

        AS-->>MA: Access Token\nRefresh Token

        Note over MA: Cache token\nper user + MCP
    end

    %% ==========================================================
    %% CALL MCP WITH TOKEN
    %% ==========================================================

    rect rgb(245,255,245)
        Note over MA,MCP2: Call MCP with OAuth Token

        MA->>MCP2: Tool Call\nAuthorization: Bearer access_token

        MCP2->>MCP2: Verify JWT Signature
        MCP2->>MCP2: Validate issuer
        MCP2->>MCP2: Validate audience
        MCP2->>MCP2: Validate scopes

        MCP2-->>MA: Tool Result

        MA-->>User: Final Answer
    end

    %% ==========================================================
    %% TOKEN REFRESH FLOW
    %% ==========================================================

    rect rgb(255,245,245)
        Note over MA,AS: Access Token Expired

        MA->>MCP2: Tool Call + Expired Token

        MCP2-->>MA: 401 Invalid Token

        MA->>AS: Refresh Token Grant

        AS-->>MA: New Access Token

        MA->>MCP2: Retry Tool Call\nwith New Token

        MCP2-->>MA: Tool Result

        MA-->>User: Final Answer
    end