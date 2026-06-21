multi-agents/
├── cmd/
│   └── main.go             # Điểm chạy chính (Khởi tạo Orchestrator)
├── internal/
│   ├── agent/              # Định nghĩa các Agent chuyên biệt
│   │   ├── agent.go        # Interface chung cho Agent
│   │   ├── search_agent.go # Agent chuyên về tra cứu (sử dụng hybrid_search tool)
│   │   ├── business_agent.go # Agent chuyên về nghiệp vụ (stylist, branch tools)
│   │   └── manager_agent.go  # (Tùy chọn) Agent quản lý/giám sát
│   ├── orchestrator/       # Bộ não điều phối
│   │   ├── router.go       # Quyết định câu hỏi này nên chuyển cho Agent nào
│   │   └── workflow.go     # Quản lý luồng làm việc giữa các Agent (nếu cần)
│   ├── mcp/                # MCP Client 
│   │   └── client.go       # Kết nối và gọi các tools từ mcp-server
│   ├── provider/           # LLM Providers (bọc lại OpenAI, Groq...)
│   │   └── llm.go          # Tái sử dụng logic gọi LLM từ dự án cũ
│   └── memory/             # Quản lý ngữ cảnh/hội thoại
│       └── session.go      # Lưu trữ context chung để các Agent cùng hiểu
├── go.mod
└── config.yaml             # Cấu hình danh sách Agent và LLM
