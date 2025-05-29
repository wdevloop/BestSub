# Context
Filename: Task-NetQuality.md
Created On: $(date +"%Y-%m-%d %H:%M:%S")
Created By: AI
Associated Protocol: RIPER-5 + Multidimensional + Agent Protocol

# Task Description
Integrate network quality checks from `net.sh` into the Go project. Specifically:
1. Measure TCP latency to various provinces for China Telecom, China Unicom, and China Mobile.
2. Determine backbone routes to China for major ISPs in key cities (Beijing, Shanghai, Guangzhou) for TCP and UDP.
3. Measure TCP latency to various international nodes.
Update `proxy/info/info.go` structs (`NetLatencyInfo`, `NetInfo`) and implement logic in `proxy/checker/` to populate these structs.

# Project Overview
The project is `BestSub`, a proxy utility. The goal is to add network quality information to `ProxyInfo`.

---
*The following sections are maintained by the AI during protocol execution*
---

# Analysis (Populated by RESEARCH mode)
The primary goal is to integrate functionalities from `net.sh` and implicitly parts of `ip.sh` (due to shared utility functions or data sources) into the Go application.

**Key functionalities to implement:**
1.  **China Latency Tests:** Measure TCP latency to various Chinese provinces for China Telecom, China Unicom, and China Mobile. This corresponds to `get_latency` and `get_operator_latency` in `net.sh`. It relies on `ping_test` (using `mtr`) and parsing its output. The list of provinces and their test targets comes from `province.json` and `target.json`.
2.  **Return Route Detection:** Determine backbone routes from the proxy server to key Chinese cities (Beijing, Shanghai, Guangzhou) for major ISPs (Telecom, Unicom, Mobile) using TCP and UDP. This corresponds to `get_route`, `get_route_mode`, and related functions like `mtr_test` or `nexttrace_test` in `net.sh`. It relies on `provider.json` for ISP information.
3.  **International Latency Tests:** Measure TCP latency to various international nodes. This corresponds to `iperf_test` in `net.sh`. It relies on `target.json` for international test nodes.

**Data Structures in `info.go` to be populated:**
*   `NetLatencyInfo`:
    *   `ChinaTelecom map[string]uint16`: Province name to latency.
    *   `ChinaUnicom map[string]uint16`: Province name to latency.
    *   `ChinaMobile map[string]uint16`: Province name to latency.
    *   `International map[string]uint16`: International node name/description to latency.
*   `NetInfo`:
    *   `Route map[string]string`: Key (e.g., "Beijing-ChinaUnicom-TCP") to route description string.

**Dependencies from Shell Scripts:**
*   **External Tools:** `mtr`, `nexttrace`, `ping`, `iperf3`, `jq`, `curl`, `bc`. These will need to be either:
    *   Replaced with Go native libraries (e.g., Go's `net/http` for curl, custom ping logic, or mtr/nexttrace parsing).
    *   Executed as external commands from Go (less ideal for portability and control).
*   **Reference Files (to be downloaded at startup from a mirror):**
    *   The Go application should download these files at startup to a temporary directory if they don't exist or if an update is forced.
    *   The base URL for downloads will be `https://ghfast.top/`.
    *   Original base GitHub raw content path: `https://raw.githubusercontent.com/`.
    *   The specific repository and paths are:
        *   `xykt/NetQuality/main/ref/`:
            *   `province.json`: (User-mentioned, e.g., `https://ghfast.top/https://raw.githubusercontent.com/xykt/NetQuality/main/ref/province.json`)
            *   `iso3166.json`: (e.g., `https://ghfast.top/https://raw.githubusercontent.com/xykt/NetQuality/main/ref/iso3166.json`)
            *   `provider.json`: (e.g., `https://ghfast.top/https://raw.githubusercontent.com/xykt/NetQuality/main/ref/provider.json`)
            *   `target.json`: (e.g., `https://ghfast.top/https://raw.githubusercontent.com/xykt/NetQuality/main/ref/target.json`)
            *   `useragent.txt`: (e.g., `https://ghfast.top/https://raw.githubusercontent.com/xykt/NetQuality/main/ref/useragent.txt`)
            *   `cookies.txt`: (e.g., `https://ghfast.top/https://raw.githubusercontent.com/xykt/NetQuality/main/ref/cookies.txt`) - Also present in IPQuality
            *   `iata-icao.csv`: (e.g., `https://ghfast.top/https://raw.githubusercontent.com/xykt/NetQuality/main/ref/iata-icao.csv`) - Also present in IPQuality
        *   `xykt/IPQuality/main/ref/`: (Some files might be duplicated or similar to NetQuality's; prefer NetQuality versions if functionally identical for this task, or download as needed if distinct).
            *   `iso3166.json`: (Likely identical to NetQuality's version)
            *   `cookies.txt`: (Likely identical to NetQuality's version)
            *   `iata-icao.csv`: (Likely identical to NetQuality's version)
            *   `dnsbl.list`: (Used by `ip.sh` for email blacklist checks, likely not relevant for `NetQuality` task)
            *   `sponsor.ans`, `ad1.ans`: (Likely not relevant for `NetQuality` task)

**Parsing Logic:**
*   Complex parsing of `mtr`, `nexttrace`, and `ping` output will be required.
*   JSON files (`province.json`, `target.json`, `provider.json`, `iso3166.json`) will need to be unmarshalled into Go structs.

**Key Naming and Struct Design:**
*   The user mentioned that keys in maps (e.g., `NetLatencyInfo.ChinaTelecom`, `NetInfo.Route`) might need a specific format or a new struct design. This will be addressed in the INNOVATE/PLAN phase.
    *   For `NetInfo.Route`, the example "Beijing-ChinaUnicom-TCP" suggests a format of "City-ISP-Protocol".

**Constraints/Considerations:**
*   Portability: Avoid direct shell command execution if possible.
*   Performance: Network operations (ping, mtr, download) can be slow. Implementations should be concurrent where feasible and have appropriate timeouts.
*   Error Handling: Robust error handling for network failures, parsing errors, etc.
*   Temporary File Management: Securely and cleanly manage downloaded temporary files.

[Code investigation results, key files, dependencies, constraints, etc.]

# Proposed Solution (Populated by INNOVATE mode)
The chosen solution is **Scheme 1: Basic download and access**.
The Go application will perform the following at startup:
1.  Define a list of required resource files (JSON, TXT, CSV) and their remote paths (relative to `https://raw.githubusercontent.com/`).
2.  Determine a dedicated application temporary directory (e.g., `os.TempDir()/BestSub/ref_data/`).
3.  For each required file:
    a.  Construct the full download URL using the mirror prefix `https://ghfast.top/` followed by the original `https://raw.githubusercontent.com/` path.
    b.  Check if the file already exists in the local temporary directory.
    c.  If the file does not exist, download it from the constructed mirror URL and save it to the local temporary directory.
4.  Runtime access: Other parts of the application will read these resource files directly from their known local temporary paths.

This approach prioritizes simplicity for the initial implementation. Files are only downloaded if they are missing locally. No automatic update mechanism for existing files is included in this scheme.

# Implementation Plan (Generated by PLAN mode)

**Part A: Resource Management (Steps 1-7 Completed)**
Implementation Checklist:
1. Update `Task-NetQuality.md`: Add the chosen solution ('Scheme 1: Basic download and access') to the "Proposed Solution" section. *(Status: Success)*
2. Create `proxy/checker/resource_manager.go`. *(Status: Success)*
3. In `resource_manager.go`: Define `ResourceFile` struct (`Name string`, `RemotePath string`, `LocalPath string`) and a list of `ResourceFile` instances for all required files (`province.json`, `iso3166.json`, `provider.json`, `target.json`, `useragent.txt`, `cookies.txt`, `iata-icao.csv` from `xykt/NetQuality/main/ref/`). *(Status: Success)*
4. In `resource_manager.go`: Implement `getAppTempDir() (string, error)` function to get/create `os.TempDir()/BestSub/ref_data/`. *(Status: Success)*
5. In `resource_manager.go`: Implement `downloadFile(filepath string, url string) error` function using HTTP GET. *(Status: Success)*
6. In `resource_manager.go`: Implement `EnsureResourceFiles() error` function. This function will:
    a. Call `getAppTempDir()`.
    b. Iterate through the predefined resource list.
    c. For each resource, construct its local path and full download URL (using `https://ghfast.top/https://raw.githubusercontent.com/` prefix).
    d. Check if the local file exists using `os.Stat()`.
    e. If the file does not exist, call `downloadFile()` to download it.
    f. Log download activities and handle errors (return error for critical file download failures). *(Status: Success)*
7. In `resource_manager.go`: Implement `GetResourcePath(name string) (string, error)` to return the full local path for a given resource name by looking it up in the predefined list and combining with the app temp dir. *(Status: Success)*

**Part B: Network Quality Core Logic (New Plan)**
8. In `main.go` (or an appropriate initialization function, e.g., in an `init()` block or early in `main()`): Call `checker.EnsureResourceFiles()`. Handle any returned error (e.g., log and exit if critical resources fail to download).
9. Create a new file: `proxy/checker/net_quality_checker.go`. This file will house the core logic for network quality checks.
10. In `proxy/checker/net_quality_checker.go`: Define Go structs to parse data from JSON resource files:
    a.  Struct for `province.json` items (e.g., `ProvinceInfo { Code, Short, Name, Targets ... }`).
    b.  Struct for `target.json` items (e.g., `TargetInfo { ID, Name, Host, Type ... }`), specifically for international latency tests and potentially China latency test IP/host details if not directly in `province.json`.
    c.  Struct for `provider.json` items (e.g., `ProviderInfo { Name, ASN, ... }`) for route analysis.
11. In `proxy/checker/net_quality_checker.go`: Implement a generic helper function `loadResourceJSON(resourceName string, targetStructPointer interface{}) error`. This function will:
    a. Call `checker.GetResourcePath(resourceName)` to get the local file path.
    b. Read the file content.
    c. Use `json.Unmarshal` to parse the content into `targetStructPointer`.
12. In `proxy/checker/net_quality_checker.go`: Implement `getChinaLatency(proxy *info.Proxy, client *http.Client, baseTimeout time.Duration) (telecomLatencies map[string]uint16, unicomLatencies map[string]uint16, mobileLatencies map[string]uint16, err error)`. This function will:
    a. Load `province.json` using `loadResourceJSON`.
    b. Iterate through provinces. For each province and for each ISP (China Telecom, China Unicom, China Mobile):
        i.  Determine the target host (e.g., from `province.json` directly, or by looking up in a structure loaded from `target.json`). The format is often like `<province_code>-<isp_code>-v4.ip.zstaticcdn.com`.
        ii. Perform a TCP latency test to this host. This might involve:
            *   Option 1 (Preferred for Go-native): Using a Go library for ICMP ping (if available and suitable for TCP-like latency) or a custom TCP dial and handshake timing.
            *   Option 2 (Fallback): Executing `mtr -"$IPV" --tcp -P 80 -c 1 -f 100 -C -G 1 -s 1400 <target_host>` (adjusting for IPv4/IPv6 based on proxy) via `os/exec` and parsing the output (e.g., the average latency from the summary line). The timeout for `mtr` should be derived from `baseTimeout`.
        iii. Store the latency (uint16, in milliseconds) in the respective map (province name as key).
13. In `proxy/checker/net_quality_checker.go`: Implement `getInternationalLatency(proxy *info.Proxy, client *http.Client, baseTimeout time.Duration) (internationalLatencies map[string]uint16, err error)`. This function will:
    a. Load `target.json` (filter for international latency test targets).
    b. Iterate through international targets.
    c. Perform TCP latency test similar to `getChinaLatency` (Option 1 or 2).
    d. Store latency (uint16) in the map (target name/description as key).
14. In `proxy/checker/net_quality_checker.go`: Implement `getReturnRoutes(proxy *info.Proxy, client *http.Client, baseTimeout time.Duration) (routeInfo map[string]string, err error)`. This function will:
    a. Load `province.json` (for key cities like Beijing, Shanghai, Guangzhou) and `provider.json`.
    b. For each key city (e.g., Beijing, Shanghai, Guangzhou), ISP (Telecom, Unicom, Mobile), and protocol (TCP, UDP):
        i.  Determine target host (e.g., `bj-ct-v4.ip.zstaticcdn.com`).
        ii. Execute a route measurement command. This will likely involve `os/exec`:
            *   `nexttrace -p 80 -q 8 -"$IPV" --<protocol> --raw --psize 1400 <target_host>` OR
            *   `mtr -"$IPV" --<protocol> --no-dns -y 0 -P 80 -c 8 -C -Z 1 -G 1 -s 1400 <target_host>` (parsing the CSV output for hops, ASNs, IPs).
        iii. Parse the output to determine the route characteristics (e.g., first hop AS outside China, entry point China ASN like CN2, 4837, CMI). The logic from `net.sh`'s `nexttrace_test` or `mtr_test` functions regarding ASN identification (e.g., 4134, 4837, 58453, 9929, 4809) will need to be replicated.
        iv. Format the route string and store it in the map. The key should be like "Beijing-ChinaUnicom-TCP".
15. In `proxy/checker/net_quality_checker.go`: Create a public function, e.g., `CheckNetQualityInfo(p *info.Proxy, httpClient *http.Client, generalTimeout time.Duration) error`. This function will:
    a. Call `getChinaLatency`, `getInternationalLatency`, and `getReturnRoutes`.
    b. Populate `p.Info.Net.Latency` (with ChinaTelecom, ChinaUnicom, ChinaMobile, International maps) and `p.Info.Net.Route`.
    c. Handle and aggregate errors.
16. In `proxy/checker/checker.go` (or wherever individual proxy checks are orchestrated, likely in a function that populates `ProxyInfo`):
    a. Call `checker.CheckNetQualityInfo(&proxy.Info, proxy.Client, /* relevant timeout */)`.
    b. Ensure `proxy.Client` is passed correctly if needed by the latency/route functions, or they create their own with appropriate proxy settings.
17. Write unit tests for parsing functions and individual checker logic where feasible.
18. Review and update `Task-NetQuality.md` "Analysis" section if new insights or complexities arise during the implementation of this core logic.
19. Update `Task-NetQuality.md` "Current Execution Step" to reflect the start of Part B.

# Current Execution Step (Updated by EXECUTE mode when starting a step)
> Currently executing: "9. Create a new file: `proxy/checker/net_quality_checker.go`. This file will house the core logic for network quality checks."

# Task Progress (Appended by EXECUTE mode after each step completion)
*   $(date +"%Y-%m-%d %H:%M:%S")
    *   Step: 1. Update `Task-NetQuality.md`: Add the chosen solution ('Scheme 1: Basic download and access') to the "Proposed Solution" section. *(Status: Success)*
    *   Modifications:
        - `Task-NetQuality.md`: Updated "Proposed Solution", "Implementation Plan", and "Current Execution Step" sections.
    *   Change Summary: Task file updated with chosen solution and detailed plan.
    *   Reason: Executing plan step 1.
    *   Blockers: None
    *   User Confirmation Status: Success
*   $(date +"%Y-%m-%d %H:%M:%S")
    *   Step: 2. Create `proxy/checker/resource_manager.go`.
    *   Modifications:
        - `proxy/checker/resource_manager.go`: Created new file with package declaration.
    *   Change Summary: Created placeholder file for resource manager.
    *   Reason: Executing plan step 2.
    *   Blockers: None
    *   User Confirmation Status: Success
*   $(date +"%Y-%m-%d %H:%M:%S")
    *   Step: 3. In `resource_manager.go`: Define `ResourceFile` struct (`Name string`, `RemotePath string`, `LocalPath string`) and a list of `ResourceFile` instances for all required files (`province.json`, `iso3166.json`, `provider.json`, `target.json`, `useragent.txt`, `cookies.txt`, `iata-icao.csv` from `xykt/NetQuality/main/ref/`).
    *   Modifications:
        - `proxy/checker/resource_manager.go`: Added `ResourceFile` struct, `requiredResourceFiles` variable, and related constants.
    *   Change Summary: Defined data structures for managing resource files.
    *   Reason: Executing plan step 3.
    *   Blockers: None
    *   User Confirmation Status: Success
*   $(date +"%Y-%m-%d %H:%M:%S")
    *   Step: 4. In `resource_manager.go`: Implement `getAppTempDir() (string, error)` function...
    *   Modifications:
        - `proxy/checker/resource_manager.go`: Implemented `getAppTempDir`.
    *   Change Summary: Added function to get/create application-specific temporary directory.
    *   Reason: Executing plan step 4.
    *   Blockers: None
    *   User Confirmation Status: Pending Confirmation
*   $(date +"%Y-%m-%d %H:%M:%S")
    *   Step: 5. In `resource_manager.go`: Implement `downloadFile(filepath string, url string) error` function...
    *   Modifications:
        - `proxy/checker/resource_manager.go`: Implemented `downloadFile`.
    *   Change Summary: Added function to download a file via HTTP GET.
    *   Reason: Executing plan step 5.
    *   Blockers: None
    *   User Confirmation Status: Pending Confirmation
*   $(date +"%Y-%m-%d %H:%M:%S")
    *   Step: 6. In `resource_manager.go`: Implement `EnsureResourceFiles() error` function...
    *   Modifications:
        - `proxy/checker/resource_manager.go`: Implemented `EnsureResourceFiles`.
    *   Change Summary: Added core logic to check and download resource files if missing.
    *   Reason: Executing plan step 6.
    *   Blockers: None
    *   User Confirmation Status: Pending Confirmation
*   $(date +"%Y-%m-%d %H:%M:%S")
    *   Step: 7. In `resource_manager.go`: Implement `GetResourcePath(name string) (string, error)` function...
    *   Modifications:
        - `proxy/checker/resource_manager.go`: Implemented `GetResourcePath` and removed unused `runtime` import.
    *   Change Summary: Added function to retrieve local path of a managed resource file.
    *   Reason: Executing plan step 7.
    *   Blockers: None
    *   User Confirmation Status: Success (Implicitly by user moving to next phase)
*   $(date +"%Y-%m-%d %H:%M:%S")
    *   Step: 8. In `main.go` (or an appropriate initialization function, e.g., in an `init()` block or early in `main()`): Call `checker.EnsureResourceFiles()`. Handle any returned error (e.g., log and exit if critical resources fail to download).
    *   Modifications:
        - `main.go`: Added call to `checker.EnsureResourceFiles()` in `app.Initialize()` with basic error logging.
    *   Change Summary: Integrated resource file initialization into the application startup sequence.
    *   Reason: Executing plan step 8.
    *   Blockers: None
    *   User Confirmation Status: Pending Confirmation
*   $(date +"%Y-%m-%d %H:%M:%S")
    *   Step: 9. Create a new file: `proxy/checker/net_quality_checker.go`. This file will house the core logic for network quality checks.
    *   Modifications:
        - `proxy/checker/net_quality_checker.go`: Created new file with package declaration and placeholder comment.
    *   Change Summary: Created the Go file for network quality checking logic.
    *   Reason: Executing plan step 9.
    *   Blockers: None
    *   User Confirmation Status: Pending Confirmation

# Final Review (Populated by REVIEW mode)
[Summary of implementation compliance assessment against the final plan, whether unreported deviations were found] 