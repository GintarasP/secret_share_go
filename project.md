"One-Time Secret Sharing" service with End-to-End Encryption (E2EE).

Project Requirements:

    Security First: Always use the crypto/aes and crypto/cipher packages. We will implement AES-256-GCM.

    Zero Knowledge: Ensure the server logic is designed so that the encryption/decryption keys are never stored in plain text on the server.

    Burn-on-Read: Implement a strict "one-time" retrieval logic using Go's concurrency-safe sync.Map or a similar in-memory store.

    Clean Code: Use idiomatic Go (proper error handling, no panic, clear naming conventions).

Instructions for the Agent:

    When I ask for code, provide small, modular functions (e.g., separate the encryption logic from the HTTP handlers).

    Explain why certain security choices are made (e.g., why we use a nonce or why GCM is preferred over CBC).

    Suggest ways to use Go's standard library before recommending third-party packages.

    Help me write unit tests for the encryption/decryption logic to ensure data integrity.

Current Task: "Let’s start by setting up the project structure. Can you provide the Go code for a robust Encrypt and Decrypt function using AES-GCM, and explain how we should handle the 32-byte key?"