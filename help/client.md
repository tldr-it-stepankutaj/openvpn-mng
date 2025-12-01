# OpenVPN Client Integration Guide

This document describes how to create a Go client that integrates OpenVPN server with OpenVPN Manager API. The client replaces PHP scripts for authentication, connection handling, and firewall rules generation.

## Table of Contents

- [Overview](#overview)
- [Architecture](#architecture)
- [Authentication Methods](#authentication-methods)
- [API Endpoints Used](#api-endpoints-used)
- [Go Client Implementation](#go-client-implementation)
- [OpenVPN Server Configuration](#openvpn-server-configuration)
- [Firewall Integration](#firewall-integration)
- [Deployment](#deployment)
- [Server-Side Changes for API Token](#server-side-changes-for-api-token)

---

## Overview

The OpenVPN Manager client handles four main functions:

1. **Authentication** (`auth-user-pass-verify`) - Validates user credentials against the API
2. **Client Connect** (`client-connect`) - Configures client IP, pushes routes based on group membership
3. **Client Disconnect** (`client-disconnect`) - Records session end and traffic statistics
4. **Firewall Rules** - Generates nftables or iptables rules based on user-network assignments

---

## Architecture

```
┌─────────────────┐     ┌─────────────────┐     ┌─────────────────┐
│  OpenVPN Server │────▶│   Go Client     │────▶│ OpenVPN Manager │
│                 │     │   (scripts)     │     │     API         │
└─────────────────┘     └─────────────────┘     └─────────────────┘
                               │
                               ▼
                        ┌─────────────────┐
                        │ nftables/iptables│
                        │   (firewall)    │
                        └─────────────────┘
```

---

## Authentication Methods

The client supports two authentication methods for API access:

### Option 1: API Token (Recommended)

Uses a static API token configured in the OpenVPN Manager server. This is the **recommended** approach because:

- Token is not a real user account (cannot login to web UI)
- No password expiration issues
- Simpler configuration
- Better security isolation

**Client config:**
```yaml
api:
  base_url: "http://127.0.0.1:8080"
  token: "your-api-token-from-server-config"
  timeout: 10s
```

**Server config (config.yaml):**
```yaml
api:
  enabled: true
  vpn_token: "generate-a-secure-random-token-here"
```

### Option 2: Service Account (Legacy)

Uses a dedicated user account with ADMIN role. Less secure because:

- Account could be used to login to web UI
- Password needs to be managed
- Appears in user list

**Client config:**
```yaml
api:
  base_url: "http://127.0.0.1:8080"
  username: "vpn-service"
  password: "secure-service-password"
  timeout: 10s
```

---

## API Endpoints Used

The client uses these API endpoints:

| Endpoint | Method | Purpose | Auth Required |
|----------|--------|---------|---------------|
| `/api/v1/auth/login` | POST | Validate VPN user credentials | No |
| `/api/v1/vpn-auth/authenticate` | POST | Validate VPN user (token auth) | VPN Token |
| `/api/v1/vpn-auth/users` | GET | List all active users (firewall) | VPN Token |
| `/api/v1/vpn-auth/users/{id}` | GET | Get user by ID | VPN Token |
| `/api/v1/vpn-auth/users/{id}/routes` | GET | Get user's allowed networks | VPN Token |
| `/api/v1/vpn-auth/users/by-username/{username}` | GET | Get user by username | VPN Token |
| `/api/v1/vpn-auth/sessions` | POST | Create VPN session | VPN Token |
| `/api/v1/vpn-auth/sessions/{id}/disconnect` | PUT | End VPN session | VPN Token |

---

## Go Client Implementation

### Project Structure

```
openvpn-client/
├── cmd/
│   ├── login/main.go           # auth-user-pass-verify script
│   ├── connect/main.go         # client-connect script
│   ├── disconnect/main.go      # client-disconnect script
│   └── firewall/main.go        # firewall rules generator
├── internal/
│   ├── api/
│   │   └── client.go           # API client
│   ├── config/
│   │   └── config.go           # Configuration
│   ├── firewall/
│   │   ├── firewall.go         # Firewall interface
│   │   ├── nftables.go         # nftables implementation
│   │   └── iptables.go         # iptables implementation
│   └── utils/
│       └── cidr.go             # CIDR to netmask conversion
├── config.yaml
├── go.mod
└── go.sum
```

### Configuration (config.yaml)

```yaml
api:
  base_url: "http://127.0.0.1:8080"
  # Option 1: API Token (recommended)
  token: "your-api-token-from-server-config"
  # Option 2: Service account (legacy)
  # username: "vpn-service"
  # password: "secure-service-password"
  timeout: 10s

openvpn:
  # Directory for temporary session files
  session_dir: "/var/run/openvpn"

firewall:
  # Firewall type: "nftables" or "iptables"
  type: "nftables"

  # nftables settings
  nftables:
    rules_file: "/etc/nftables.d/vpn-users.nft"
    reload_command: "/usr/sbin/nft -f /etc/sysconfig/nftables.conf"

  # iptables settings
  iptables:
    chain_name: "VPN_USERS"
    rules_file: "/etc/iptables.d/vpn-users.rules"
    reload_command: "/usr/sbin/iptables-restore < /etc/iptables.d/vpn-users.rules"
```

### go.mod

```go
module openvpn-client

go 1.22

require (
    gopkg.in/yaml.v3 v3.0.1
)
```

### internal/config/config.go

```go
package config

import (
    "os"
    "time"

    "gopkg.in/yaml.v3"
)

type Config struct {
    API      APIConfig      `yaml:"api"`
    OpenVPN  OpenVPNConfig  `yaml:"openvpn"`
    Firewall FirewallConfig `yaml:"firewall"`
}

type APIConfig struct {
    BaseURL  string        `yaml:"base_url"`
    Token    string        `yaml:"token"`    // API token (recommended)
    Username string        `yaml:"username"` // Service account (legacy)
    Password string        `yaml:"password"` // Service account (legacy)
    Timeout  time.Duration `yaml:"timeout"`
}

type OpenVPNConfig struct {
    SessionDir string `yaml:"session_dir"`
}

type FirewallConfig struct {
    Type     string          `yaml:"type"` // "nftables" or "iptables"
    NFTables NFTablesConfig  `yaml:"nftables"`
    IPTables IPTablesConfig  `yaml:"iptables"`
}

type NFTablesConfig struct {
    RulesFile     string `yaml:"rules_file"`
    ReloadCommand string `yaml:"reload_command"`
}

type IPTablesConfig struct {
    ChainName     string `yaml:"chain_name"`
    RulesFile     string `yaml:"rules_file"`
    ReloadCommand string `yaml:"reload_command"`
}

// UseToken returns true if API token authentication should be used
func (c *APIConfig) UseToken() bool {
    return c.Token != ""
}

func Load(path string) (*Config, error) {
    data, err := os.ReadFile(path)
    if err != nil {
        return nil, err
    }

    var cfg Config
    if err := yaml.Unmarshal(data, &cfg); err != nil {
        return nil, err
    }

    // Defaults
    if cfg.API.Timeout == 0 {
        cfg.API.Timeout = 10 * time.Second
    }
    if cfg.OpenVPN.SessionDir == "" {
        cfg.OpenVPN.SessionDir = "/var/run/openvpn"
    }
    if cfg.Firewall.Type == "" {
        cfg.Firewall.Type = "nftables"
    }

    return &cfg, nil
}
```

### internal/api/client.go

```go
package api

import (
    "bytes"
    "encoding/json"
    "fmt"
    "io"
    "net/http"
    "net/url"
    "time"

    "openvpn-client/internal/config"
)

type Client struct {
    baseURL    string
    httpClient *http.Client
    token      string      // JWT token (from service account login)
    apiToken   string      // Static API token (from config)
}

// Response types
type LoginResponse struct {
    Token     string       `json:"token"`
    ExpiresAt time.Time    `json:"expires_at"`
    User      UserResponse `json:"user"`
}

type UserResponse struct {
    ID        string     `json:"id"`
    Username  string     `json:"username"`
    FirstName string     `json:"first_name"`
    LastName  string     `json:"last_name"`
    Email     string     `json:"email"`
    Role      string     `json:"role"`
    IsActive  bool       `json:"is_active"`
    ValidFrom *time.Time `json:"valid_from"`
    ValidTo   *time.Time `json:"valid_to"`
    VpnIP     string     `json:"vpn_ip"`
}

type UserListResponse struct {
    Users      []UserResponse `json:"users"`
    Total      int64          `json:"total"`
    Page       int            `json:"page"`
    PageSize   int            `json:"page_size"`
    TotalPages int            `json:"total_pages"`
}

type GroupWithNetworks struct {
    ID          string    `json:"id"`
    Name        string    `json:"name"`
    Description string    `json:"description"`
    Networks    []Network `json:"networks"`
}

type GroupsResponse struct {
    Groups []GroupWithNetworks `json:"groups"`
}

type Network struct {
    ID          string `json:"id"`
    Name        string `json:"name"`
    CIDR        string `json:"cidr"`
    Description string `json:"description"`
}

type VpnSession struct {
    ID             string     `json:"id"`
    UserID         string     `json:"user_id"`
    VpnIP          string     `json:"vpn_ip"`
    ClientIP       string     `json:"client_ip"`
    ConnectedAt    time.Time  `json:"connected_at"`
    DisconnectedAt *time.Time `json:"disconnected_at"`
    BytesReceived  int64      `json:"bytes_received"`
    BytesSent      int64      `json:"bytes_sent"`
}

type VpnAuthRequest struct {
    Username string `json:"username"`
    Password string `json:"password"`
}

type VpnAuthResponse struct {
    Valid   bool         `json:"valid"`
    User    UserResponse `json:"user"`
    Message string       `json:"message,omitempty"`
}

type ErrorResponse struct {
    Error   string `json:"error"`
    Message string `json:"message"`
}

func NewClient(cfg *config.APIConfig) *Client {
    c := &Client{
        baseURL: cfg.BaseURL,
        httpClient: &http.Client{
            Timeout: cfg.Timeout,
        },
    }

    if cfg.UseToken() {
        c.apiToken = cfg.Token
    }

    return c
}

// SetAPIToken sets the static API token for authentication
func (c *Client) SetAPIToken(token string) {
    c.apiToken = token
}

// Authenticate gets a JWT token using service account credentials (legacy)
func (c *Client) Authenticate(username, password string) error {
    body := map[string]string{
        "username": username,
        "password": password,
    }

    resp, err := c.doRequest("POST", "/api/v1/auth/login", body, false)
    if err != nil {
        return err
    }
    defer resp.Body.Close()

    if resp.StatusCode != http.StatusOK {
        return c.parseError(resp)
    }

    var loginResp LoginResponse
    if err := json.NewDecoder(resp.Body).Decode(&loginResp); err != nil {
        return err
    }

    c.token = loginResp.Token
    return nil
}

// ValidateVpnUser validates VPN user credentials using API token
func (c *Client) ValidateVpnUser(username, password string) (*VpnAuthResponse, error) {
    body := VpnAuthRequest{
        Username: username,
        Password: password,
    }

    // Use VPN-specific endpoint if using API token
    endpoint := "/api/v1/vpn-auth/authenticate"
    if c.apiToken == "" {
        // Fallback to regular login for legacy service account
        endpoint = "/api/v1/auth/login"
    }

    resp, err := c.doRequest("POST", endpoint, body, c.apiToken != "")
    if err != nil {
        return nil, err
    }
    defer resp.Body.Close()

    if c.apiToken != "" {
        // VPN auth endpoint returns VpnAuthResponse
        if resp.StatusCode != http.StatusOK {
            return &VpnAuthResponse{Valid: false, Message: "authentication failed"}, nil
        }

        var authResp VpnAuthResponse
        if err := json.NewDecoder(resp.Body).Decode(&authResp); err != nil {
            return nil, err
        }
        return &authResp, nil
    } else {
        // Legacy login endpoint
        if resp.StatusCode != http.StatusOK {
            return &VpnAuthResponse{Valid: false, Message: "authentication failed"}, nil
        }

        var loginResp LoginResponse
        if err := json.NewDecoder(resp.Body).Decode(&loginResp); err != nil {
            return nil, err
        }
        return &VpnAuthResponse{Valid: true, User: loginResp.User}, nil
    }
}

// GetUserByUsername finds a user by username
func (c *Client) GetUserByUsername(username string) (*UserResponse, error) {
    // Use VPN-specific endpoint if using API token
    if c.apiToken != "" {
        resp, err := c.doRequest("GET", "/api/v1/vpn-auth/users/by-username/"+url.PathEscape(username), nil, true)
        if err != nil {
            return nil, err
        }
        defer resp.Body.Close()

        if resp.StatusCode != http.StatusOK {
            return nil, c.parseError(resp)
        }

        var user UserResponse
        if err := json.NewDecoder(resp.Body).Decode(&user); err != nil {
            return nil, err
        }
        return &user, nil
    }

    // Legacy: search users
    params := url.Values{}
    params.Set("search", username)
    params.Set("page_size", "100")

    resp, err := c.doRequest("GET", "/api/v1/users?"+params.Encode(), nil, true)
    if err != nil {
        return nil, err
    }
    defer resp.Body.Close()

    if resp.StatusCode != http.StatusOK {
        return nil, c.parseError(resp)
    }

    var listResp UserListResponse
    if err := json.NewDecoder(resp.Body).Decode(&listResp); err != nil {
        return nil, err
    }

    for _, user := range listResp.Users {
        if user.Username == username {
            return &user, nil
        }
    }

    return nil, fmt.Errorf("user not found: %s", username)
}

// GetUserRoutes gets user's allowed networks (routes)
func (c *Client) GetUserRoutes(userID string) ([]Network, error) {
    // Use VPN-specific endpoint if using API token
    endpoint := "/api/v1/vpn-auth/users/" + userID + "/routes"
    if c.apiToken == "" {
        endpoint = "/api/v1/users/" + userID + "/groups"
    }

    resp, err := c.doRequest("GET", endpoint, nil, true)
    if err != nil {
        return nil, err
    }
    defer resp.Body.Close()

    if resp.StatusCode != http.StatusOK {
        return nil, c.parseError(resp)
    }

    if c.apiToken != "" {
        // VPN endpoint returns networks directly
        var networks []Network
        if err := json.NewDecoder(resp.Body).Decode(&networks); err != nil {
            return nil, err
        }
        return networks, nil
    }

    // Legacy: parse groups response
    var groupsResp GroupsResponse
    if err := json.NewDecoder(resp.Body).Decode(&groupsResp); err != nil {
        return nil, err
    }

    // Flatten networks from all groups
    networkMap := make(map[string]Network)
    for _, group := range groupsResp.Groups {
        for _, network := range group.Networks {
            networkMap[network.CIDR] = network
        }
    }

    networks := make([]Network, 0, len(networkMap))
    for _, n := range networkMap {
        networks = append(networks, n)
    }
    return networks, nil
}

// GetAllActiveUsers gets all active users for firewall rules
func (c *Client) GetAllActiveUsers() ([]UserResponse, error) {
    // Use VPN-specific endpoint if using API token
    if c.apiToken != "" {
        resp, err := c.doRequest("GET", "/api/v1/vpn-auth/users", nil, true)
        if err != nil {
            return nil, err
        }
        defer resp.Body.Close()

        if resp.StatusCode != http.StatusOK {
            return nil, c.parseError(resp)
        }

        var users []UserResponse
        if err := json.NewDecoder(resp.Body).Decode(&users); err != nil {
            return nil, err
        }
        return users, nil
    }

    // Legacy: paginated user list
    var allUsers []UserResponse
    page := 1
    pageSize := 100

    for {
        params := url.Values{}
        params.Set("page", fmt.Sprintf("%d", page))
        params.Set("page_size", fmt.Sprintf("%d", pageSize))
        params.Set("is_active", "true")

        resp, err := c.doRequest("GET", "/api/v1/users?"+params.Encode(), nil, true)
        if err != nil {
            return nil, err
        }

        var listResp UserListResponse
        if err := json.NewDecoder(resp.Body).Decode(&listResp); err != nil {
            resp.Body.Close()
            return nil, err
        }
        resp.Body.Close()

        allUsers = append(allUsers, listResp.Users...)

        if page >= listResp.TotalPages {
            break
        }
        page++
    }

    return allUsers, nil
}

// CreateSession creates a new VPN session
func (c *Client) CreateSession(userID, vpnIP, clientIP string) (*VpnSession, error) {
    body := map[string]interface{}{
        "user_id":      userID,
        "vpn_ip":       vpnIP,
        "client_ip":    clientIP,
        "connected_at": time.Now().UTC().Format(time.RFC3339),
    }

    resp, err := c.doRequest("POST", "/api/v1/vpn-auth/sessions", body, true)
    if err != nil {
        return nil, err
    }
    defer resp.Body.Close()

    if resp.StatusCode != http.StatusCreated {
        return nil, c.parseError(resp)
    }

    var session VpnSession
    if err := json.NewDecoder(resp.Body).Decode(&session); err != nil {
        return nil, err
    }

    return &session, nil
}

// DisconnectSession ends a VPN session
func (c *Client) DisconnectSession(sessionID string, bytesReceived, bytesSent int64) error {
    body := map[string]interface{}{
        "disconnected_at":   time.Now().UTC().Format(time.RFC3339),
        "bytes_received":    bytesReceived,
        "bytes_sent":        bytesSent,
        "disconnect_reason": "USER_REQUEST",
    }

    resp, err := c.doRequest("PUT", "/api/v1/vpn-auth/sessions/"+sessionID+"/disconnect", body, true)
    if err != nil {
        return err
    }
    defer resp.Body.Close()

    if resp.StatusCode != http.StatusOK {
        return c.parseError(resp)
    }

    return nil
}

func (c *Client) doRequest(method, path string, body interface{}, auth bool) (*http.Response, error) {
    var bodyReader io.Reader
    if body != nil {
        jsonBody, err := json.Marshal(body)
        if err != nil {
            return nil, err
        }
        bodyReader = bytes.NewReader(jsonBody)
    }

    req, err := http.NewRequest(method, c.baseURL+path, bodyReader)
    if err != nil {
        return nil, err
    }

    req.Header.Set("Content-Type", "application/json")

    if auth {
        if c.apiToken != "" {
            // Use X-VPN-Token header for API token auth
            req.Header.Set("X-VPN-Token", c.apiToken)
        } else if c.token != "" {
            // Use Bearer token for JWT auth (legacy)
            req.Header.Set("Authorization", "Bearer "+c.token)
        }
    }

    return c.httpClient.Do(req)
}

func (c *Client) parseError(resp *http.Response) error {
    body, _ := io.ReadAll(resp.Body)

    var errResp ErrorResponse
    if err := json.Unmarshal(body, &errResp); err == nil && errResp.Message != "" {
        return fmt.Errorf("%s: %s", errResp.Error, errResp.Message)
    }

    return fmt.Errorf("API error: %s (status %d)", string(body), resp.StatusCode)
}
```

### internal/firewall/firewall.go

```go
package firewall

import (
    "openvpn-client/internal/api"
    "openvpn-client/internal/config"
)

// UserWithNetworks represents a user with their allowed networks
type UserWithNetworks struct {
    Username string
    VpnIP    string
    Networks []string
}

// Firewall is the interface for firewall rule generators
type Firewall interface {
    // GenerateRules generates firewall rules for the given users
    GenerateRules(users []UserWithNetworks) string
    // GetRulesFile returns the path to the rules file
    GetRulesFile() string
    // GetReloadCommand returns the command to reload firewall rules
    GetReloadCommand() string
}

// New creates a new firewall based on configuration
func New(cfg *config.FirewallConfig) Firewall {
    switch cfg.Type {
    case "iptables":
        return NewIPTables(&cfg.IPTables)
    default:
        return NewNFTables(&cfg.NFTables)
    }
}

// CollectUserNetworks collects networks for all users from API
func CollectUserNetworks(client *api.Client, users []api.UserResponse) ([]UserWithNetworks, error) {
    var result []UserWithNetworks

    for _, user := range users {
        if user.VpnIP == "" {
            continue
        }

        routes, err := client.GetUserRoutes(user.ID)
        if err != nil {
            // Skip user on error, don't fail entire operation
            continue
        }

        networks := make([]string, 0)
        for _, route := range routes {
            // Skip default route for firewall rules
            if route.CIDR == "0.0.0.0/0" || route.CIDR == "0/0" {
                continue
            }
            networks = append(networks, route.CIDR)
        }

        if len(networks) > 0 {
            result = append(result, UserWithNetworks{
                Username: user.Username,
                VpnIP:    user.VpnIP,
                Networks: networks,
            })
        }
    }

    return result, nil
}
```

### internal/firewall/nftables.go

```go
package firewall

import (
    "fmt"
    "sort"
    "strings"

    "openvpn-client/internal/config"
)

type NFTables struct {
    rulesFile     string
    reloadCommand string
}

func NewNFTables(cfg *config.NFTablesConfig) *NFTables {
    return &NFTables{
        rulesFile:     cfg.RulesFile,
        reloadCommand: cfg.ReloadCommand,
    }
}

func (n *NFTables) GenerateRules(users []UserWithNetworks) string {
    var rules strings.Builder
    rules.WriteString("# Auto-generated VPN user rules (nftables)\n")
    rules.WriteString("# Do not edit manually - changes will be overwritten\n\n")

    for _, user := range users {
        // Sort networks for consistent output
        sort.Strings(user.Networks)

        rules.WriteString(fmt.Sprintf("# %s\n", user.Username))
        rules.WriteString(fmt.Sprintf("ip saddr %s ip daddr { %s } accept\n",
            user.VpnIP,
            strings.Join(user.Networks, ", ")))
    }

    return rules.String()
}

func (n *NFTables) GetRulesFile() string {
    return n.rulesFile
}

func (n *NFTables) GetReloadCommand() string {
    return n.reloadCommand
}
```

### internal/firewall/iptables.go

```go
package firewall

import (
    "fmt"
    "sort"
    "strings"

    "openvpn-client/internal/config"
)

type IPTables struct {
    chainName     string
    rulesFile     string
    reloadCommand string
}

func NewIPTables(cfg *config.IPTablesConfig) *IPTables {
    chainName := cfg.ChainName
    if chainName == "" {
        chainName = "VPN_USERS"
    }
    return &IPTables{
        chainName:     chainName,
        rulesFile:     cfg.RulesFile,
        reloadCommand: cfg.ReloadCommand,
    }
}

func (i *IPTables) GenerateRules(users []UserWithNetworks) string {
    var rules strings.Builder
    rules.WriteString("# Auto-generated VPN user rules (iptables)\n")
    rules.WriteString("# Do not edit manually - changes will be overwritten\n\n")
    rules.WriteString("*filter\n")

    // Create/flush chain
    rules.WriteString(fmt.Sprintf(":%s - [0:0]\n", i.chainName))
    rules.WriteString(fmt.Sprintf("-F %s\n", i.chainName))

    for _, user := range users {
        // Sort networks for consistent output
        sort.Strings(user.Networks)

        rules.WriteString(fmt.Sprintf("# %s\n", user.Username))
        for _, network := range user.Networks {
            rules.WriteString(fmt.Sprintf("-A %s -s %s -d %s -j ACCEPT\n",
                i.chainName, user.VpnIP, network))
        }
    }

    rules.WriteString("COMMIT\n")
    return rules.String()
}

func (i *IPTables) GetRulesFile() string {
    return i.rulesFile
}

func (i *IPTables) GetReloadCommand() string {
    return i.reloadCommand
}
```

### internal/utils/cidr.go

```go
package utils

import (
    "fmt"
    "net"
    "strings"
)

// CIDRToNetmask converts CIDR notation to IP and netmask
// Example: "192.168.1.0/24" -> "192.168.1.0 255.255.255.0"
func CIDRToNetmask(cidr string) (string, error) {
    // If no slash, treat as single IP
    if !strings.Contains(cidr, "/") {
        ip := net.ParseIP(cidr)
        if ip == nil {
            return "", fmt.Errorf("invalid IP: %s", cidr)
        }
        return cidr + " 255.255.255.255", nil
    }

    // Parse CIDR
    _, ipNet, err := net.ParseCIDR(cidr)
    if err != nil {
        return "", err
    }

    // Get network address and mask
    ip := ipNet.IP.String()
    mask := net.IP(ipNet.Mask).String()

    return ip + " " + mask, nil
}
```

### cmd/login/main.go (auth-user-pass-verify)

```go
package main

import (
    "fmt"
    "os"
    "strings"

    "openvpn-client/internal/api"
    "openvpn-client/internal/config"
)

func main() {
    if len(os.Args) < 2 {
        fmt.Fprintln(os.Stderr, "Error: password file path not provided")
        os.Exit(1)
    }

    // Read credentials file
    data, err := os.ReadFile(os.Args[1])
    if err != nil {
        fmt.Fprintf(os.Stderr, "Error reading auth file: %v\n", err)
        os.Exit(1)
    }

    lines := strings.Split(string(data), "\n")
    if len(lines) < 2 {
        fmt.Fprintln(os.Stderr, "Error: invalid auth file format")
        os.Exit(1)
    }

    username := strings.TrimSpace(lines[0])
    password := strings.TrimSpace(lines[1])

    // Load configuration
    cfg, err := config.Load("/etc/openvpn/client/config.yaml")
    if err != nil {
        fmt.Fprintf(os.Stderr, "Error loading config: %v\n", err)
        os.Exit(1)
    }

    // Create API client
    client := api.NewClient(&cfg.API)

    // Validate credentials
    authResp, err := client.ValidateVpnUser(username, password)
    if err != nil {
        fmt.Fprintf(os.Stderr, "Authentication error for %s: %v\n", username, err)
        os.Exit(1)
    }

    if !authResp.Valid {
        fmt.Fprintf(os.Stderr, "Authentication failed for %s: %s\n", username, authResp.Message)
        os.Exit(1)
    }

    fmt.Printf("User %s successfully authenticated (ID: %s)\n", authResp.User.Username, authResp.User.ID)
    os.Exit(0)
}
```

### cmd/connect/main.go (client-connect)

```go
package main

import (
    "fmt"
    "os"
    "path/filepath"
    "strings"

    "openvpn-client/internal/api"
    "openvpn-client/internal/config"
    "openvpn-client/internal/utils"
)

func main() {
    if len(os.Args) < 2 {
        fmt.Fprintln(os.Stderr, "Error: config file path not provided")
        os.Exit(1)
    }

    configFilePath := os.Args[1]

    // Get environment variables from OpenVPN
    commonName := os.Getenv("common_name")
    if commonName == "" {
        fmt.Fprintln(os.Stderr, "Error: common_name not set")
        os.Exit(1)
    }

    trustedIP := os.Getenv("trusted_ip")
    trustedPort := os.Getenv("trusted_port")
    remoteIP := os.Getenv("ifconfig_pool_remote_ip")

    // Load configuration
    cfg, err := config.Load("/etc/openvpn/client/config.yaml")
    if err != nil {
        fmt.Fprintf(os.Stderr, "Error loading config: %v\n", err)
        os.Exit(1)
    }

    // Create API client
    client := api.NewClient(&cfg.API)

    // Authenticate if using legacy service account
    if !cfg.API.UseToken() {
        if err := client.Authenticate(cfg.API.Username, cfg.API.Password); err != nil {
            fmt.Fprintf(os.Stderr, "API authentication failed: %v\n", err)
            os.Exit(1)
        }
    }

    // Get user by username
    user, err := client.GetUserByUsername(commonName)
    if err != nil {
        fmt.Fprintf(os.Stderr, "User not found: %s - %v\n", commonName, err)
        os.Exit(1)
    }

    // Build config file content
    var configContent strings.Builder

    // Set static VPN IP if configured
    vpnIP := remoteIP
    if user.VpnIP != "" {
        vpnIP = user.VpnIP
        configContent.WriteString(fmt.Sprintf("ifconfig-push %s 255.255.255.0\n", user.VpnIP))
    }

    // Get user's routes
    routes, err := client.GetUserRoutes(user.ID)
    if err != nil {
        fmt.Fprintf(os.Stderr, "Error getting user routes: %v\n", err)
        os.Exit(1)
    }

    // Check for default route and collect networks
    hasDefaultRoute := false
    var networks []string

    for _, route := range routes {
        if route.CIDR == "0.0.0.0/0" || route.CIDR == "0/0" {
            hasDefaultRoute = true
            continue
        }
        networks = append(networks, route.CIDR)
    }

    // Push routes
    if hasDefaultRoute {
        configContent.WriteString("push \"redirect-gateway def1\"\n")
    } else {
        for _, cidr := range networks {
            route, err := utils.CIDRToNetmask(cidr)
            if err != nil {
                fmt.Fprintf(os.Stderr, "Warning: invalid CIDR %s: %v\n", cidr, err)
                continue
            }
            configContent.WriteString(fmt.Sprintf("push \"route %s\"\n", route))
        }
    }

    // Write config file
    if err := os.WriteFile(configFilePath, []byte(configContent.String()), 0644); err != nil {
        fmt.Fprintf(os.Stderr, "Error writing config file: %v\n", err)
        os.Exit(1)
    }

    // Create VPN session
    session, err := client.CreateSession(user.ID, vpnIP, trustedIP)
    if err != nil {
        fmt.Fprintf(os.Stderr, "Warning: could not create session: %v\n", err)
    } else {
        // Save session ID for disconnect script
        sessionFile := filepath.Join(cfg.OpenVPN.SessionDir, fmt.Sprintf("session-%s", commonName))
        sessionData := fmt.Sprintf("%s\n%s\n%s", session.ID, trustedIP, trustedPort)
        if err := os.WriteFile(sessionFile, []byte(sessionData), 0600); err != nil {
            fmt.Fprintf(os.Stderr, "Warning: could not save session file: %v\n", err)
        }
    }

    fmt.Printf("Client %s connected with IP %s\n", commonName, vpnIP)
    os.Exit(0)
}
```

### cmd/disconnect/main.go (client-disconnect)

```go
package main

import (
    "fmt"
    "os"
    "path/filepath"
    "strconv"
    "strings"

    "openvpn-client/internal/api"
    "openvpn-client/internal/config"
)

func main() {
    commonName := os.Getenv("common_name")
    if commonName == "" {
        fmt.Fprintln(os.Stderr, "Error: common_name not set")
        os.Exit(1)
    }

    bytesReceived, _ := strconv.ParseInt(os.Getenv("bytes_received"), 10, 64)
    bytesSent, _ := strconv.ParseInt(os.Getenv("bytes_sent"), 10, 64)

    // Load configuration
    cfg, err := config.Load("/etc/openvpn/client/config.yaml")
    if err != nil {
        fmt.Fprintf(os.Stderr, "Error loading config: %v\n", err)
        os.Exit(1)
    }

    // Read session file
    sessionFile := filepath.Join(cfg.OpenVPN.SessionDir, fmt.Sprintf("session-%s", commonName))
    data, err := os.ReadFile(sessionFile)
    if err != nil {
        fmt.Fprintf(os.Stderr, "Warning: session file not found: %v\n", err)
        os.Exit(0)
    }

    lines := strings.Split(string(data), "\n")
    if len(lines) < 1 {
        fmt.Fprintln(os.Stderr, "Warning: invalid session file")
        os.Exit(0)
    }

    sessionID := strings.TrimSpace(lines[0])

    // Create API client
    client := api.NewClient(&cfg.API)

    // Authenticate if using legacy service account
    if !cfg.API.UseToken() {
        if err := client.Authenticate(cfg.API.Username, cfg.API.Password); err != nil {
            fmt.Fprintf(os.Stderr, "API authentication failed: %v\n", err)
            os.Exit(0)
        }
    }

    // End session
    if err := client.DisconnectSession(sessionID, bytesReceived, bytesSent); err != nil {
        fmt.Fprintf(os.Stderr, "Warning: could not end session: %v\n", err)
    }

    // Remove session file
    os.Remove(sessionFile)

    fmt.Printf("Client %s disconnected (received: %d, sent: %d)\n", commonName, bytesReceived, bytesSent)
    os.Exit(0)
}
```

### cmd/firewall/main.go (firewall rules generator)

```go
package main

import (
    "fmt"
    "os"
    "os/exec"

    "openvpn-client/internal/api"
    "openvpn-client/internal/config"
    "openvpn-client/internal/firewall"
)

func main() {
    // Load configuration
    cfg, err := config.Load("/etc/openvpn/client/config.yaml")
    if err != nil {
        fmt.Fprintf(os.Stderr, "Error loading config: %v\n", err)
        os.Exit(1)
    }

    // Create API client
    client := api.NewClient(&cfg.API)

    // Authenticate if using legacy service account
    if !cfg.API.UseToken() {
        if err := client.Authenticate(cfg.API.Username, cfg.API.Password); err != nil {
            fmt.Fprintf(os.Stderr, "API authentication failed: %v\n", err)
            os.Exit(1)
        }
    }

    // Get all active users
    users, err := client.GetAllActiveUsers()
    if err != nil {
        fmt.Fprintf(os.Stderr, "Error getting users: %v\n", err)
        os.Exit(1)
    }

    // Collect networks for each user
    usersWithNetworks, err := firewall.CollectUserNetworks(client, users)
    if err != nil {
        fmt.Fprintf(os.Stderr, "Error collecting networks: %v\n", err)
        os.Exit(1)
    }

    // Create firewall generator
    fw := firewall.New(&cfg.Firewall)

    // Generate rules
    newRules := fw.GenerateRules(usersWithNetworks)

    // Check if rules changed
    oldRules, _ := os.ReadFile(fw.GetRulesFile())
    if string(oldRules) == newRules {
        fmt.Println("Firewall rules unchanged")
        os.Exit(0)
    }

    // Write new rules
    if err := os.WriteFile(fw.GetRulesFile(), []byte(newRules), 0644); err != nil {
        fmt.Fprintf(os.Stderr, "Error writing rules file: %v\n", err)
        os.Exit(1)
    }

    // Reload firewall
    cmd := exec.Command("sh", "-c", fw.GetReloadCommand())
    if output, err := cmd.CombinedOutput(); err != nil {
        fmt.Fprintf(os.Stderr, "Error reloading firewall: %v\n%s\n", err, output)
        os.Exit(1)
    }

    fmt.Printf("Firewall rules updated (%d users, type: %s)\n", len(usersWithNetworks), cfg.Firewall.Type)
    os.Exit(0)
}
```

---

## Building the Client

```bash
cd openvpn-client

# Build all binaries
go build -o /usr/local/bin/openvpn-login ./cmd/login
go build -o /usr/local/bin/openvpn-connect ./cmd/connect
go build -o /usr/local/bin/openvpn-disconnect ./cmd/disconnect
go build -o /usr/local/bin/openvpn-firewall ./cmd/firewall

# Set permissions
chmod 755 /usr/local/bin/openvpn-*
```

---

## OpenVPN Server Configuration

Update your OpenVPN server configuration:

```conf
### OPENVPN SERVER CONFIG ###

mode server
tls-server
local 51.77.51.151
port 993
proto tcp
dev tun
float
topology subnet

ca /etc/openvpn/victoriatech/ca.crt
cert /etc/openvpn/victoriatech/vpn.victoriatech.cz.crt
key /etc/openvpn/victoriatech/vpn.victoriatech.cz.key
dh /etc/openvpn/victoriatech/dh.pem
crl-verify /etc/openvpn/victoriatech/crl.pem

server 10.90.90.0 255.255.255.0

push "topology subnet"

keepalive 10 120

tls-auth /etc/openvpn/victoriatech/ta.key 0

cipher AES-128-GCM
ncp-ciphers AES-128-GCM:AES-256-GCM
compress stub-v2
push "compress stub-v2"

user openvpn
group openvpn

persist-key
persist-tun

status /etc/openvpn/victoriatech/openvpn-status.log
log-append /var/log/openvpn_victoriatech.log

verb 3

management localhost 7505

# Authentication via API
username-as-common-name
auth-user-pass-verify /usr/local/bin/openvpn-login via-file
client-connect /usr/local/bin/openvpn-connect
client-disconnect /usr/local/bin/openvpn-disconnect
script-security 2
reneg-sec 0
```

---

## Firewall Integration

### NFTables Configuration

**Main config (/etc/sysconfig/nftables.conf):**

```nft
#!/usr/sbin/nft -f

flush ruleset

table inet filter {
    chain input {
        type filter hook input priority 0; policy drop;
        ct state established,related accept
        iif lo accept
        tcp dport 993 accept
        tcp dport 22 accept
    }

    chain forward {
        type filter hook forward priority 0; policy drop;
        ct state established,related accept
        # VPN user rules (auto-generated)
        include "/etc/nftables.d/vpn-users.nft"
    }

    chain output {
        type filter hook output priority 0; policy accept;
    }
}

table inet nat {
    chain postrouting {
        type nat hook postrouting priority 100;
        ip saddr 10.90.90.0/24 masquerade
    }
}
```

**Client config:**

```yaml
firewall:
  type: "nftables"
  nftables:
    rules_file: "/etc/nftables.d/vpn-users.nft"
    reload_command: "/usr/sbin/nft -f /etc/sysconfig/nftables.conf"
```

### IPTables Configuration

**Main config (/etc/sysconfig/iptables):**

```bash
*filter
:INPUT DROP [0:0]
:FORWARD DROP [0:0]
:OUTPUT ACCEPT [0:0]
:VPN_USERS - [0:0]

-A INPUT -m state --state ESTABLISHED,RELATED -j ACCEPT
-A INPUT -i lo -j ACCEPT
-A INPUT -p tcp --dport 993 -j ACCEPT
-A INPUT -p tcp --dport 22 -j ACCEPT

-A FORWARD -m state --state ESTABLISHED,RELATED -j ACCEPT
-A FORWARD -i tun0 -j VPN_USERS

COMMIT

*nat
:POSTROUTING ACCEPT [0:0]
-A POSTROUTING -s 10.90.90.0/24 -j MASQUERADE
COMMIT
```

**Client config:**

```yaml
firewall:
  type: "iptables"
  iptables:
    chain_name: "VPN_USERS"
    rules_file: "/etc/iptables.d/vpn-users.rules"
    reload_command: "iptables-restore -n < /etc/iptables.d/vpn-users.rules"
```

### Cron job for firewall updates

```bash
# /etc/cron.d/openvpn-firewall
*/5 * * * * root /usr/local/bin/openvpn-firewall >> /var/log/openvpn-firewall.log 2>&1
```

---

## Deployment

### 1. Create directories

```bash
mkdir -p /etc/openvpn/client
mkdir -p /var/run/openvpn
mkdir -p /etc/nftables.d  # or /etc/iptables.d for iptables
chown openvpn:openvpn /var/run/openvpn
```

### 2. Create configuration file

**With API Token (recommended):**

```bash
cat > /etc/openvpn/client/config.yaml << 'EOF'
api:
  base_url: "http://127.0.0.1:8080"
  token: "your-vpn-token-from-server-config"
  timeout: 10s

openvpn:
  session_dir: "/var/run/openvpn"

firewall:
  type: "nftables"
  nftables:
    rules_file: "/etc/nftables.d/vpn-users.nft"
    reload_command: "/usr/sbin/nft -f /etc/sysconfig/nftables.conf"
EOF

chmod 600 /etc/openvpn/client/config.yaml
```

### 3. Configure VPN token in OpenVPN Manager

Add to server's `config.yaml`:

```yaml
api:
  enabled: true
  vpn_token: "generate-a-secure-random-token-here"
```

Generate a secure token:

```bash
openssl rand -hex 32
```

### 4. Build and install binaries

```bash
cd openvpn-client
go build -o /usr/local/bin/openvpn-login ./cmd/login
go build -o /usr/local/bin/openvpn-connect ./cmd/connect
go build -o /usr/local/bin/openvpn-disconnect ./cmd/disconnect
go build -o /usr/local/bin/openvpn-firewall ./cmd/firewall
chmod 755 /usr/local/bin/openvpn-*
```

### 5. Initialize firewall rules

```bash
touch /etc/nftables.d/vpn-users.nft  # or /etc/iptables.d/vpn-users.rules
/usr/local/bin/openvpn-firewall
```

### 6. Restart OpenVPN

```bash
systemctl restart openvpn@victoriatech
```

---

## Server-Side Changes for API Token

The OpenVPN Manager server now includes built-in support for API token authentication. The following features have been implemented:

### 1. Configuration (config.yaml)

```yaml
api:
  enabled: true
  swagger_enabled: true
  # VPN client API token (used instead of service account)
  vpn_token: "your-secure-random-token"
```

Environment variable: `API_VPN_TOKEN`

### 2. Implemented VPN-specific endpoints

These endpoints are available at `/api/v1/vpn-auth/*` and require the `X-VPN-Token` header:

| Endpoint | Method | Description |
|----------|--------|-------------|
| `POST /api/v1/vpn-auth/authenticate` | Validate VPN user credentials |
| `GET /api/v1/vpn-auth/users` | List all active users with VPN IP |
| `GET /api/v1/vpn-auth/users/{id}` | Get user by ID |
| `GET /api/v1/vpn-auth/users/{id}/routes` | Get user's allowed networks |
| `GET /api/v1/vpn-auth/users/by-username/{username}` | Get user by username |
| `POST /api/v1/vpn-auth/sessions` | Create VPN session |
| `PUT /api/v1/vpn-auth/sessions/{id}/disconnect` | End VPN session |

**Note:** These endpoints are only available when `vpn_token` is configured.

### 3. VPN Token Middleware

The middleware validates the `X-VPN-Token` header:

- Location: `internal/middleware/vpn_token.go`
- Header name: `X-VPN-Token`
- Returns 401 if token is missing or invalid
- Returns 503 if VPN token is not configured on server

### 4. Benefits of API Token vs Service Account

| Aspect | API Token | Service Account |
|--------|-----------|-----------------|
| Web UI access | No | Yes (security risk) |
| Password management | None | Required |
| User list visibility | Not shown | Shown as user |
| Audit trail | Separate | Mixed with users |
| Token rotation | Simple config change | Password change + client update |
| Permissions | VPN endpoints only | Full ADMIN access |

---

## Troubleshooting

### Check authentication

```bash
# Create test auth file
echo -e "testuser\ntestpassword" > /tmp/auth.txt
/usr/local/bin/openvpn-login /tmp/auth.txt
echo $?  # Should be 0 for success
```

### Check API connectivity

```bash
# With API token - list all VPN users
curl -H "X-VPN-Token: your-token" http://127.0.0.1:8080/api/v1/vpn-auth/users

# Get user by username
curl -H "X-VPN-Token: your-token" http://127.0.0.1:8080/api/v1/vpn-auth/users/by-username/testuser

# With service account (legacy)
TOKEN=$(curl -s -X POST http://127.0.0.1:8080/api/v1/auth/login \
  -H "Content-Type: application/json" \
  -d '{"username":"vpn-service","password":"your-password"}' | jq -r .token)
curl -H "Authorization: Bearer $TOKEN" http://127.0.0.1:8080/api/v1/users
```

### View logs

```bash
tail -f /var/log/openvpn_victoriatech.log
tail -f /var/log/openvpn-firewall.log
```

---

## Migration from PHP Scripts

| PHP Script | Go Binary | Function |
|------------|-----------|----------|
| `login.php` | `openvpn-login` | User authentication |
| `connect.php` | `openvpn-connect` | Client config, session start |
| `disconnect.php` | `openvpn-disconnect` | Session end, traffic stats |
| `gen-nftables.php` | `openvpn-firewall` | Firewall rules generation |

Key differences:
- Uses REST API instead of direct database access
- Passwords validated by API (bcrypt) instead of SHA256
- Session tracking via VPN sessions API
- Compiled binary instead of interpreted PHP
- Supports both nftables and iptables
- API token authentication (no service account needed)
