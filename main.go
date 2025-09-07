package main

import (
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"syscall"
	"unsafe"
	"time"

	"bufio"

	"github.com/fatih/color"
	"go.zoe.im/injgo"
)

type File struct {
	URL  string
	Name string
}

// Constants for DLL injection because i always forget what this does
const (
	PROCESS_ALL_ACCESS     = 0x1F0FFF
	MEM_COMMIT             = 0x1000
	MEM_RESERVE            = 0x2000
	PAGE_EXECUTE_READWRITE = 0x40
	STD_OUTPUT_HANDLE      = -11
)

// Function prototypes for DLL injection
var (
	procVirtualAllocEx     = syscall.NewLazyDLL("kernel32.dll").NewProc("VirtualAllocEx")
	procWriteProcessMemory = syscall.NewLazyDLL("kernel32.dll").NewProc("WriteProcessMemory")
	procCreateRemoteThread = syscall.NewLazyDLL("kernel32.dll").NewProc("CreateRemoteThread")
)

func main() {
	// redirect
	if !fileExists("redirect.json") {
		file := []byte(`{ "name": "Starfall.dll", "download": "https://github.com/Ducki67/Fortnite-redirect-dlls/raw/refs/heads/main/starfall.dll%20(all)/Starfall.dll" }`)
		err := ioutil.WriteFile("redirect.json", file, 0644)
		if err != nil {
			panic(err)
		}
	}

	fileData, err := ioutil.ReadFile("redirect.json")
	if err != nil {
		panic(err)
	}

	var jsonData map[string]interface{}
	err = json.Unmarshal(fileData, &jsonData)
	if err != nil {
		panic(err)
	}

	// console
	if !fileExists("console.json") {
		file := []byte(`{ "name": "Console.dll", "download": "https://github.com/Ducki67/Fortnite-redirect-dlls/raw/refs/heads/main/console.dll" }`)  // console dll for s24 OR s1
		err := ioutil.WriteFile("console.json", file, 0644)
		if err != nil {
			panic(err)
		}
	}

	fileData, err = ioutil.ReadFile("console.json")
	if err != nil {
		panic(err)
	}

	var consoleData map[string]interface{}
	err = json.Unmarshal(fileData, &consoleData)
	if err != nil {
		panic(err)
	}

	starfallName, ok := jsonData["name"].(string)
	if !ok {
		panic("Invalid JSON structure: 'name' key is missing or not a string")
	}
	starfallDownload, ok := jsonData["download"].(string)
	if !ok {
		panic("Invalid JSON structure: 'download' key is missing or not a string")
	}
	if !strings.HasSuffix(starfallName, ".dll") {
		starfallName += ".dll"
	}

	consoleName, ok := consoleData["name"].(string)
	if !ok {
		panic("Invalid JSON structure: 'name' key is missing or not a string")
	}
	consoleDownload, ok := consoleData["download"].(string)
	if !ok {
		panic("Invalid JSON structure: 'download' key is missing or not a string")
	}
	if !strings.HasSuffix(consoleName, ".dll") {
		consoleName += ".dll"
	}

	fileList := []File{
		{URL: starfallDownload, Name: starfallName},
		{URL: consoleDownload, Name: consoleName},
		/**/
        {URL: "https://cdn.discordapp.com/attachments/958139296936783892/1000707724507623424/FortniteClient-Win64-Shipping_BE.exe", Name: "FortniteClient-Win64-Shipping_BE.exe"},
		{URL: "https://cdn.discordapp.com/attachments/958139296936783892/1000707724818006046/FortniteLauncher.exe", Name: "FortniteLauncher.exe"},
		
		/*
		/// 24.20 /// 
		{URL: "https://cdn.discordapp.com/attachments/1413166239160860723/1413201825514913935/FortniteClient-Win64-Shipping_EAC.exe", Name: "FortniteClient-Win64-Shipping_EAC.exe"},
	    {URL: "https://cdn.discordapp.com/attachments/1413166239160860723/1413166314243096686/FortniteLauncher.exe", Name: "FortniteLauncher.exe"},
		{URL: "https://cdn.discordapp.com/attachments/1413166239160860723/1413166448129474600/FortniteClient-Win64-Shipping_BE.exe", Name: "FortniteClient-Win64-Shipping_BE.exe"},
		{URL: "https://cdn.discordapp.com/attachments/1413166239160860723/1413167014737875075/FortniteClient-Win64-Shipping.exe", Name: "FortniteClient-Win64-Shipping.exe"}, // for s24.20 exe
		*/
	}
	localappdata := fmt.Sprintf("%s/AppData/Local/LightyearV2", os.Getenv("USERPROFILE"))
	color.Magenta(`
   ┌───────────────────────────┐
   │ ╦  ┬┌─┐┬ ┬┌┬┐┬ ┬┌─┐┌─┐┬─┐ │
   │ ║  ││ ┬├─┤ │ └┬┘├┤ ├─┤├┬┘ │
   │ ╩═╝┴└─┘┴ ┴ ┴  ┴ └─┘┴ ┴┴└─ │
   └─┬───────────────────────┬─┘
   ┌─┘      LightyearV2      └─┐
   │     Fortnite Launcher     │
   └───────────────────────────┘
	`)

	if !folderExists(localappdata) {
		createFolder(localappdata)
	}

	for _, file := range fileList {
		filename := filepath.Base(file.Name)
		localPath := filepath.Join(localappdata, filename)

		if !fileExists(localPath) {
			color.Blue("Downloading missing file %s", filename)
			err := downloadFile(file.URL, localPath)
			if err != nil {
				panic(err)
			}
		}
	}

	//Ask for input
	var input string
	color.New(color.FgHiCyan).Println(`
   ┌───────────────────────────────────────┐
   │ ► Options:                     ━ □ X  │
   │ Select a option to continue           │
   ├───────────────────────────────────────┤
   │ ► 1. Start Fortnite                   │
   │ ► 2. Add/Change Fortnite path         │
   │ ► 3. Add/Change email and password    │
   │ ► 4. Show Info tab (expermental)      │
   │ ► 5. Build Downlaod tab (expermental) │
   └───────────────────────────────────────┘
	`)

	fmt.Scanln(&input)

	switch input {
	case "1":
		go runFortnite(localappdata)
		var inout string
		color.Red("[OPTION] Press enter to exit / close the Launcher")
		fmt.Scanln(&inout)
	case "2":
		changePath(localappdata)
		main()
	case "3":
		color.Yellow("Please enter your email")
		var email string
		fmt.Scanln(&email)
		color.Yellow("Please enter your password")
		var password string
		fmt.Scanln(&password)

		emailFile, err := os.Create(localappdata + "/email.txt")
		if err != nil {
			panic(err)
		}
		defer emailFile.Close()
		fmt.Fprintf(emailFile, "%s", email)

		passwordFile, err := os.Create(localappdata + "/password.txt")
		if err != nil {
			panic(err)
		}
		defer passwordFile.Close()
		fmt.Fprintf(passwordFile, "%s", password)

		password = strings.Repeat("*", len(password))




    case "4":
	var input string
		color.New(color.FgHiCyan).Println(`
   ┌───────────────────────────────────────────────┐
   │ ▼ Info:                               ━ □ X   │
   │ Welcome to the Launcher info tab!             │
   ├───────────────────────────────────────────────┤
   │                                               │
   │ * Credits:                                    │
   │   - Ducki67                                   │
   │    Fortnite build test list,                  │
   │    Added options 4 and 5 to the UI,           │
   │                                               │
   │   - simplyzetax                               │
   │   Made the launcher main source,              │
   │                                               │
   │ * Fortnite Version Support:                   │
   │                                               │
   │   [+] 1.8.2-CL-3741772   (fully works)        │
   │   [+] 1.8-CL-3724489     (fully works)        │
   │   [+] 1.11-CL-3807424    (fully works)        │
   │   [X] 7.40-CL-5046157    (very broken)        │
   │   [*] 23.50 -CL-N/A      (never gonna work)   │
   │   [~] 1.smthing          (partially working)  │
   │                                               │
   │                                               │
   │                                               │
   └───────────────────────────────────────────────┘
	`)
	
	fmt.Scanln(&input)



	 case "5":
	var input string
		color.New(color.FgHiCyan).Println(`
   ╔═══════════════════════════════════════════════════════════════════╗
   ║ ▼ Build Downlaod:                                          ━ □ X  ║
   ║ Welcome to the FN Build download tab!                             ║
   ╠═══════════════════════════════════════════════════════════════════╣
   ║                                                                   ║
   ║ * If you want to download builds use (CTR + Click) in the links!  ║
   ║                                                                   ║
   ║ * Season 0 & 1                                                    ║
   ║   ┌───────┬───────────────────────────────────────────────────┐   ║
   ║   │ Build │                  Downlaod Links                   │   ║
   ║   ├───────┼───────────────────────────────────────────────────┤   ║
   ║   │ 1.7.2 │ https://public.simplyblk.xyz/1.7.2.zip            │   ║
   ║   │ 1.8   │ https://public.simplyblk.xyz/1.8.rar              │   ║
   ║   │ 1.8.1 │ https://public.simplyblk.xyz/1.8.1.rar            │   ║
   ║   │ 1.8.2 │ https://public.simplyblk.xyz/1.8.2.rar            │   ║
   ║   │ 1.9   │ https://public.simplyblk.xyz/1.9.rar              │   ║
   ║   │ 1.9.1 │ https://public.simplyblk.xyz/1.9.1.rar            │   ║
   ║   │ 1.10  │ https://public.simplyblk.xyz/1.10.rar             │   ║
   ║   └───────┴───────────────────────────────────────────────────┘   ║
   ║                                                                   ║
   ║ * Season 2                                                        ║
   ║   ┌───────┬───────────────────────────────────────────────────┐   ║
   ║   │ Build │                  Downlaod Links                   │   ║
   ║   ├───────┼───────────────────────────────────────────────────┤   ║
   ║   │ 1.11  │ https://public.simplyblk.xyz/1.11.zip             │   ║
   ║   │ 2.1   │ https://public.simplyblk.xyz/2.1.0.zip            │   ║
   ║   │ 2.2   │ https://public.simplyblk.xyz/2.2.0.rar            │   ║
   ║   │ 2.3   │ https://public.simplyblk.xyz/2.3.rar              │   ║
   ║   │ 2.4.0 │ https://public.simplyblk.xyz/2.4.0.zip            │   ║
   ║   │ 2.4.2 │ https://public.simplyblk.xyz/2.4.2.zip            │   ║
   ║   │ 2.5   │ https://public.simplyblk.xyz/2.5.0.rar            │   ║
   ║   └───────┴───────────────────────────────────────────────────┘   ║
   ║                                                                   ║
   ║ * Season 3                                                        ║
   ║   ┌───────┬───────────────────────────────────────────────────┐   ║
   ║   │ Build │                  Downlaod Links                   │   ║
   ║   ├───────┼───────────────────────────────────────────────────┤   ║
   ║   │ 3.0   │ https://public.simplyblk.xyz/3.0.zip              │   ║
   ║   │ 3.1   │ https://public.simplyblk.xyz/3.1.rar              │   ║
   ║   │ 3.1.1 │ https://public.simplyblk.xyz/3.1.1.zip            │   ║
   ║   │ 3.2   │ https://public.simplyblk.xyz/3.2.zip              │   ║
   ║   │ 3.3   │ https://public.simplyblk.xyz/3.3.rar              │   ║
   ║   │ 3.5   │ https://public.simplyblk.xyz/3.5.rar              │   ║
   ║   │ 3.6   │ https://public.simplyblk.xyz/3.6.zip              │   ║
   ║   └───────┴───────────────────────────────────────────────────┘   ║
   ║                                                                   ║
   ║ * Season 4                                                        ║
   ║   ┌───────┬───────────────────────────────────────────────────┐   ║
   ║   │ Build │                  Downlaod Links                   │   ║
   ║   ├───────┼───────────────────────────────────────────────────┤   ║
   ║   │ 4.0   │ https://public.simplyblk.xyz/4.0.zip              │   ║
   ║   │ 4.1   │ https://public.simplyblk.xyz/4.1.zip              │   ║
   ║   │ 4.2   │ https://public.simplyblk.xyz/4.2.zip              │   ║
   ║   │ 4.4   │ https://public.simplyblk.xyz/4.4.rar              │   ║
   ║   │ 4.5   │ https://public.simplyblk.xyz/4.5.rar              │   ║
   ║   └───────┴───────────────────────────────────────────────────┘   ║
   ║                                                                   ║
   ║ * Season 5                                                        ║
   ║   ┌───────┬───────────────────────────────────────────────────┐   ║
   ║   │ Build │                  Downlaod Links                   │   ║
   ║   ├───────┼───────────────────────────────────────────────────┤   ║
   ║   │ 5.0   │ https://public.simplyblk.xyz/5.00.rar             │   ║
   ║   │ 5.0.1 │ https://public.simplyblk.xyz/5.0.1.rar            │   ║
   ║   │ 5.10  │ https://public.simplyblk.xyz/5.10.rar             │   ║
   ║   │ 5.21  │ https://public.simplyblk.xyz/5.21.rar             │   ║
   ║   │ 5.30  │ https://public.simplyblk.xyz/5.30.rar             │   ║
   ║   │ 5.40  │ https://public.simplyblk.xyz/5.40.rar             │   ║
   ║   └───────┴───────────────────────────────────────────────────┘   ║
   ║                                                                   ║
   ║ * Season 6                                                        ║
   ║   ┌───────┬───────────────────────────────────────────────────┐   ║
   ║   │ Build │                  Downlaod Links                   │   ║
   ║   ├───────┼───────────────────────────────────────────────────┤   ║
   ║   │ 6.00  │ https://public.simplyblk.xyz/6.00.rar             │   ║
   ║   │ 6.01  │ https://public.simplyblk.xyz/6.01.rar             │   ║
   ║   │ 6.01.1│ https://public.simplyblk.xyz/6.1.1.rar            │   ║
   ║   │ 6.02  │ https://public.simplyblk.xyz/6.02.rar             │   ║
   ║   │ 6.02.1│ https://public.simplyblk.xyz/6.2.1.rar            │   ║
   ║   │ 6.10  │ https://public.simplyblk.xyz/6.10.rar             │   ║
   ║   │ 6.10.1│ https://public.simplyblk.xyz/6.10.1.rar           │   ║
   ║   │ 6.10.2│ https://public.simplyblk.xyz/6.10.2.rar           │   ║
   ║   │ 6.21  │ https://public.simplyblk.xyz/6.21.rar             │   ║
   ║   │ 6.22  │ https://public.simplyblk.xyz/6.22.rar             │   ║
   ║   │ 6.30  │ https://public.simplyblk.xyz/6.30.rar             │   ║
   ║   │ 6.31  │ https://public.simplyblk.xyz/6.31.rar             │   ║
   ║   └───────┴───────────────────────────────────────────────────┘   ║
   ║                                                                   ║
   ║ * Season 7                                                        ║
   ║   ┌───────┬───────────────────────────────────────────────────┐   ║
   ║   │ Build │                  Downlaod Links                   │   ║
   ║   ├───────┼───────────────────────────────────────────────────┤   ║
   ║   │ 7.00  │ https://public.simplyblk.xyz/7.00.rar             │   ║
   ║   │ 7.10  │ https://public.simplyblk.xyz/7.10.rar             │   ║
   ║   │ 7.20  │ https://public.simplyblk.xyz/7.20.rar             │   ║
   ║   │ 7.30  │ https://public.simplyblk.xyz/7.30.zip             │   ║
   ║   │ 7.40  │ https://public.simplyblk.xyz/7.40.rar             │   ║
   ║   └───────┴───────────────────────────────────────────────────┘   ║
   ║                                                                   ║
   ║ * Season 8                                                        ║
   ║   ┌───────┬───────────────────────────────────────────────────┐   ║
   ║   │ Build │                  Downlaod Links                   │   ║
   ║   ├───────┼───────────────────────────────────────────────────┤   ║
   ║   │ 8.00  │ https://public.simplyblk.xyz/8.00.zip             │   ║
   ║   │ 8.20  │ https://public.simplyblk.xyz/8.20.rar             │   ║
   ║   │ 8.30  │ https://public.simplyblk.xyz/8.30.rar             │   ║
   ║   │ 8.40  │ https://public.simplyblk.xyz/8.40.zip             │   ║
   ║   │ 8.50  │ https://public.simplyblk.xyz/8.50.zip             │   ║
   ║   │ 8.51  │ https://public.simplyblk.xyz/8.51.rar             │   ║
   ║   └───────┴───────────────────────────────────────────────────┘   ║
   ║                                                                   ║
   ║ * Season 9                                                        ║
   ║   ┌───────┬───────────────────────────────────────────────────┐   ║
   ║   │ Build │                  Downlaod Links                   │   ║
   ║   ├───────┼───────────────────────────────────────────────────┤   ║
   ║   │ 9.00  │ https://public.simplyblk.xyz/9.00.zip             │   ║
   ║   │ 9.01  │ https://public.simplyblk.xyz/9.01.zip             │   ║
   ║   │ 9.10  │ https://public.simplyblk.xyz/9.10.rar             │   ║
   ║   │ 9.21  │ https://public.simplyblk.xyz/9.21.zip             │   ║
   ║   │ 9.30  │ https://public.simplyblk.xyz/9.30.zip             │   ║
   ║   │ 9.40  │ https://public.simplyblk.xyz/9.40.zip             │   ║
   ║   │ 9.41  │ https://public.simplyblk.xyz/9.41.rar             │   ║
   ║   └───────┴───────────────────────────────────────────────────┘   ║
   ║                                                                   ║
   ║ * Season X/10                                                     ║
   ║   ┌───────┬───────────────────────────────────────────────────┐   ║
   ║   │ Build │                  Downlaod Links                   │   ║
   ║   ├───────┼───────────────────────────────────────────────────┤   ║
   ║   │ 10.00 │ https://public.simplyblk.xyz/10.00.zip            │   ║
   ║   │ 10.10 │ https://public.simplyblk.xyz/10.10.zip            │   ║
   ║   │ 10.20 │ https://public.simplyblk.xyz/10.20.zip            │   ║
   ║   │ 10.31 │ https://public.simplyblk.xyz/10.31.zip            │   ║
   ║   │ 10.40 │ https://public.simplyblk.xyz/10.40.rar            │   ║
   ║   └───────┴───────────────────────────────────────────────────┘   ║
   ║                                                                   ║
   ║ * Season 11                                                       ║
   ║   ┌───────┬───────────────────────────────────────────────────┐   ║
   ║   │ Build │                  Downlaod Links                   │   ║
   ║   ├───────┼───────────────────────────────────────────────────┤   ║
   ║   │ 11.00 │ https://public.simplyblk.xyz/11.00.zip            │   ║
   ║   │ 11.31 │ https://public.simplyblk.xyz/11.31.rar            │   ║
   ║   └───────┴───────────────────────────────────────────────────┘   ║
   ║                                                                   ║
   ║ * Season 12                                                       ║
   ║   ┌───────┬───────────────────────────────────────────────────┐   ║
   ║   │ Build │                  Downlaod Links                   │   ║
   ║   ├───────┼───────────────────────────────────────────────────┤   ║
   ║   │ 12.00 │ https://public.simplyblk.xyz/12.00.rar            │   ║
   ║   │ 12.21 │ https://public.simplyblk.xyz/12.21.zip            │   ║
   ║   │ 12.41 │ https://public.simplyblk.xyz/Fortnite%2012.41.zip │   ║
   ║   │ 12.50 │ https://public.simplyblk.xyz/12.50.zip            │   ║
   ║   │ 12.61 │ https://public.simplyblk.xyz/12.61.zip            │   ║
   ║   └───────┴───────────────────────────────────────────────────┘   ║
   ║                                                                   ║
   ║ * Season 13                                                       ║
   ║   ┌───────┬───────────────────────────────────────────────────┐   ║
   ║   │ Build │                  Downlaod Links                   │   ║
   ║   ├───────┼───────────────────────────────────────────────────┤   ║
   ║   │ 13.00 │ https://public.simplyblk.xyz/13.00.rar            │   ║
   ║   │ 13.40 │ https://public.simplyblk.xyz/13.40.zip            │   ║
   ║   └───────┴───────────────────────────────────────────────────┘   ║
   ║                                                                   ║
   ║ * Season 14                                                       ║
   ║   ┌───────┬───────────────────────────────────────────────────┐   ║
   ║   │ Build │                  Downlaod Links                   │   ║
   ║   ├───────┼───────────────────────────────────────────────────┤   ║
   ║   │ 14.00 │ https://public.simplyblk.xyz/14.00.rar            │   ║
   ║   │ 14.40 │ https://public.simplyblk.xyz/14.40.rar            │   ║
   ║   │ 14.60 │ https://public.simplyblk.xyz/14.60.rar            │   ║
   ║   └───────┴───────────────────────────────────────────────────┘   ║
   ║                                                                   ║
   ║ * Season 15                                                       ║
   ║   ┌───────┬───────────────────────────────────────────────────┐   ║
   ║   │ Build │                  Downlaod Links                   │   ║
   ║   ├───────┼───────────────────────────────────────────────────┤   ║
   ║   │ 15.20 │ https://public.simplyblk.xyz/15.20.rar            │   ║
   ║   │ 15.30 │ https://public.simplyblk.xyz/11.31.rar            │   ║
   ║   └───────┴───────────────────────────────────────────────────┘   ║
   ║                                                                   ║
   ║ * Season 16                                                       ║
   ║   ┌───────┬───────────────────────────────────────────────────┐   ║
   ║   │ Build │                  Downlaod Links                   │   ║
   ║   ├───────┼───────────────────────────────────────────────────┤   ║
   ║   │ 16.20 │ https://public.simplyblk.xyz/16.20.rar            │   ║
   ║   │ 16.30 │ https://public.simplyblk.xyz/16.30.zip            │   ║
   ║   │ 16.40 │ https://public.simplyblk.xyz/16.40.rar            │   ║
   ║   └───────┴───────────────────────────────────────────────────┘   ║
   ║                                                                   ║
   ║ * Season 17                                                       ║
   ║   ┌───────┬───────────────────────────────────────────────────┐   ║
   ║   │ Build │                  Downlaod Links                   │   ║
   ║   ├───────┼───────────────────────────────────────────────────┤   ║
   ║   │ 17.10 │ https://public.simplyblk.xyz/17.10.rar            │   ║
   ║   │ 17.30 │ https://public.simplyblk.xyz/17.30.zip            │   ║
   ║   │ 17.50 │ https://public.simplyblk.xyz/17.50.zip            │   ║
   ║   └───────┴───────────────────────────────────────────────────┘   ║
   ║                                                                   ║
   ║ * Season 18                                                       ║
   ║   ┌───────┬───────────────────────────────────────────────────┐   ║
   ║   │ Build │                  Downlaod Links                   │   ║
   ║   ├───────┼───────────────────────────────────────────────────┤   ║
   ║   │ 18.00 │ https://public.simplyblk.xyz/18.00.rar            │   ║
   ║   │ 18.30 │ https://public.simplyblk.xyz/18.30.7z             │   ║
   ║   │ 18.40 │ https://public.simplyblk.xyz/18.40.zip            │   ║
   ║   └───────┴───────────────────────────────────────────────────┘   ║
   ║                                                                   ║
   ║ * Season 19                                                       ║
   ║   ┌───────┬───────────────────────────────────────────────────┐   ║
   ║   │ Build │                  Downlaod Links                   │   ║
   ║   ├───────┼───────────────────────────────────────────────────┤   ║
   ║   │ 19.01 │ https://public.simplyblk.xyz/19.01.zip            │   ║
   ║   │ 19.10 │ https://public.simplyblk.xyz/19.10.rar            │   ║
   ║   │ 19.30 │ https://public.simplyblk.xyz/19.30.rar            │   ║
   ║   │ 19.40 │ https://public.simplyblk.xyz/19.40.7z             │   ║
   ║   └───────┴───────────────────────────────────────────────────┘   ║
   ║                                                                   ║
   ║ * Season 20                                                       ║
   ║   ┌───────┬───────────────────────────────────────────────────┐   ║
   ║   │ Build │                  Downlaod Links                   │   ║
   ║   ├───────┼───────────────────────────────────────────────────┤   ║
   ║   │ 20.00 │ https://public.simplyblk.xyz/20.00.rar            │   ║
   ║   │ 20.10 │ https://public.simplyblk.xyz/20.10.zip            │   ║
   ║   │ 20.30 │ https://public.simplyblk.xyz/20.20.zip            │   ║
   ║   │ 20.40 │ https://public.simplyblk.xyz/20.40.zip            │   ║
   ║   └───────┴───────────────────────────────────────────────────┘   ║
   ║                                                                   ║
   ║ * Season 21                                                       ║
   ║   ┌───────┬───────────────────────────────────────────────────┐   ║
   ║   │ Build │                  Downlaod Links                   │   ║
   ║   ├───────┼───────────────────────────────────────────────────┤   ║
   ║   │ 21.10 │ https://public.simplyblk.xyz/21.10.zip            │   ║
   ║   │ 21.50 │ https://public.simplyblk.xyz/21.50.zip            │   ║
   ║   │ 21.51 │ https://public.simplyblk.xyz/21.51.7z             │   ║
   ║   └───────┴───────────────────────────────────────────────────┘   ║
   ║                                                                   ║
   ║ * Season 22                                                       ║
   ║   ┌───────┬───────────────────────────────────────────────────┐   ║
   ║   │ Build │                  Downlaod Links                   │   ║
   ║   ├───────┼───────────────────────────────────────────────────┤   ║
   ║   │ 22.00 │ https://public.simplyblk.xyz/22.00.7z             │   ║
   ║   └───────┴───────────────────────────────────────────────────┘   ║
   ║                                                                   ║
   ║ * Season 21                                                       ║
   ║   ┌───────┬───────────────────────────────────────────────────┐   ║
   ║   │ Build │                  Downlaod Links                   │   ║
   ║   ├───────┼───────────────────────────────────────────────────┤   ║
   ║   │ 23.00 │ https://public.simplyblk.xyz/23.00.7z             │   ║
   ║   │ 23.10 │ https://public.simplyblk.xyz/23.10.rar            │   ║
   ║   │ 23.40 │ https://public.simplyblk.xyz/23.40.zip            │   ║
   ║   │ 23.50 │ https://public.simplyblk.xyz/23.50.zip            │   ║
   ║   └───────┴───────────────────────────────────────────────────┘   ║
   ║                                                                   ║
   ║ * If you want to download builds use (CTR + Click) in the links!  ║
   ║   Scroll up or down to find your build / version                  ║
   ║                                                                   ║
   ╚═══════════════════════════════════════════════════════════════════╝
	`)
	
	fmt.Scanln(&input)
	}
}

func runFortnite(localappdata string) {
	file, err := os.Open(localappdata + "/path.txt")
	if err != nil {
		panic(err)
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	if scanner.Scan() {
		path := scanner.Text()

		if !folderExists(path + `/FortniteGame`) {
			color.Red("Invalid path, please try again")
			main()
			return
		}

		password, err := readFile(localappdata + "/password.txt")
		if err != nil {
			panic(err)
		}

		email, err := readFile(localappdata + "/email.txt")
		if err != nil {
			panic(err)
		}
		
		args := []string{
			/*
			//OLD
			"-log",
			"-epicapp=Fortnite",
			"-epicenv=Prod",
			"-epiclocale=en-us",
			"-epicportal",
			"-skippatchcheck",
			"-nobe",
			"-fromfl=eac",
			"-fltoken=3db3ba5dcbd2e16703f3978d",*/
			//for 24.20 maybe ?
			/**/
			"-log",
			"-epicapp=Fortnite",
            "-epicenv=Prod",
            "-epiclocale=en-us",
            "-epicportal",
            "-skippatchcheck",
            "-nobe",
            "-fromfl=eac",
            "-fltoken=3db3ba5dcbd2e16703f3978d",
            "-caldera=eyJhbGciOiJFUzI1NiIsInR5cCI6IkpXVCJ9.eyJhY2NvdW50X2lkIjoiYmU5ZGE1YzJmYmVhNDQwN2IyZjQwZWJhYWQ4NTlhZDQiLCJnZW5lcmF0ZWQiOjE2Mzg3MTcyNzgsImNhbGRlcmFHdWlkIjoiMzgxMGI4NjMtMmE2NS00NDU3LTliNTgtNGRhYjNiNDgyYTg2IiwiYWNQcm92aWRlciI6IkVhc3lBbnRpQ2hlYXQiLCJub3RlcyI6IiIsImZhbGxiYWNrIjpmYWxzZX0.VAWQB67RTxhiWOxx7DBjnzDnXyyEnX7OljJm-j2d88G_WgwQ9wrE6lwMEHZHjBd1ISJdUO1UVUqkfLdU5nofBQ",
            
			fmt.Sprintf("-AUTH_LOGIN=%s", email),
			fmt.Sprintf("-AUTH_PASSWORD=%s", password),
			"-AUTH_TYPE=epic",
		}

			color.Blue("[FN PROCESS] Starting Fortnite...")
        /**/
        startLauncher(localappdata+"/FortniteLauncher.exe", nil, true)
		startLauncher(localappdata+"/FortniteClient-Win64-Shipping_BE.exe", args, true)
		

		/* // 24.20 maybe
		startLauncher(localappdata+"/FortniteLauncher.exe", args, false)
        startLauncher(localappdata+"/FortniteClient-Win64-Shipping_EAC.exe", nil, false)
        startLauncher(localappdata+"/FortniteClient-Win64-Shipping.exe", args, true)
		*/

		startShipping(path, args)
		
   

		process, err := injgo.FindProcessByName("FortniteClient-Win64-Shipping.exe")
		if err != nil {
			panic(err)
		}

		

		// Inject Starfall.dll immediately
		starfallPath := filepath.Join(localappdata, "Starfall.dll")
		err = injectDll(uint32(process.ProcessID), starfallPath)
		if err != nil {
			panic(err)
		}
		
		color.New(color.FgHiBlack).Println(`[DLL INJECTED] Starfall.dll injected!`) // color here
		color.New(color.FgHiBlack).Println(`[DLL INFO] The "Console.dll" will be injected in 30 seconds later!!`) // color here
		// Inject Console.dll after (30 seconds)
		go func() {
			time.Sleep(30 * time.Second)
			consolePath := filepath.Join(localappdata, "Console.dll")
			err := injectDll(uint32(process.ProcessID), consolePath)
			if err != nil {
				color.New(color.FgHiBlack).Println("Console.dll injection failed:", err)
				return
			}
			color.New(color.FgHiBlack).Println(`[DLL INJECTED] Console.dll injected after 30 seconds! Try using F8 or ~ to open it.`) // color here
		}()
	}
}

// injectDll function stays the same
func injectDll(processID uint32, dllPath string) error {
	hProcess, err := syscall.OpenProcess(PROCESS_ALL_ACCESS, false, processID)
	if err != nil {
		return err
	}
	defer syscall.CloseHandle(hProcess)

	dllPathAddr, _, err := procVirtualAllocEx.Call(
		uintptr(hProcess),
		0,
		uintptr(len(dllPath)),
		MEM_RESERVE|MEM_COMMIT,
		PAGE_EXECUTE_READWRITE,
	)
	if dllPathAddr == 0 {
		return err
	}

	dllPathBytes := []byte(dllPath)
	var bytesWritten uintptr
	_, _, err = procWriteProcessMemory.Call(
		uintptr(hProcess),
		dllPathAddr,
		uintptr(unsafe.Pointer(&dllPathBytes[0])),
		uintptr(len(dllPathBytes)),
		uintptr(unsafe.Pointer(&bytesWritten)),
	)
	if bytesWritten != uintptr(len(dllPathBytes)) {
		return err
	}

	kernel32, err := syscall.LoadLibrary("kernel32.dll")
	if err != nil {
		return err
	}
	defer syscall.FreeLibrary(kernel32)

	loadLibraryAddr, err := syscall.GetProcAddress(kernel32, "LoadLibraryA")
	if err != nil {
		return err
	}

	hThread, _, err := procCreateRemoteThread.Call(
		uintptr(hProcess),
		0,
		0,
		loadLibraryAddr,
		dllPathAddr,
		0,
		0,
	)
	if hThread == 0 {
		return err
	}
	defer syscall.CloseHandle(syscall.Handle(hThread))

	return nil
}

func startLauncher(path string, args []string, suspend bool) { // suspend
	color.Blue("[FN PROCESS] Starting FN launcher") // color here
	if !fileExists(path) {
		color.Red("Launcher not found, please try again")
		return
	}
	cmd := exec.Command(path)
	cmd.Start()
}

func startShipping(gamePath string, args []string) {
	color.Blue("[FN PROCESS] Starting shipping.exe")
	if !fileExists(filepath.Join(gamePath, "FortniteGame", "Binaries", "Win64", "FortniteClient-Win64-Shipping.exe")) {
		color.Red("Shipping not found, please try again")
		return
	}
	cmd := exec.Command(filepath.Join(gamePath, "FortniteGame", "Binaries", "Win64", "FortniteClient-Win64-Shipping.exe"))
	cmd.Args = append(cmd.Args, args...)
	cmd.Start()
}

func changePath(localappdata string) {
	fmt.Println("Please enter your Fortnite path")
	reader := bufio.NewReader(os.Stdin)
	path, err := reader.ReadString('\n')
	if err != nil {
		panic(err)
	}

	path = strings.TrimSpace(path)

	if !folderExists(path + `/FortniteGame`) {
		color.Red("Invalid path, please try again")
		color.White(path)
		changePath(localappdata)
	}

	file, err := os.Create(localappdata + "/path.txt")
	if err != nil {
		panic(err)
	}
	defer file.Close()

	fmt.Fprintf(file, "%s", path)

	clearConsole()
}

func readFile(filePath string) (string, error) {
	data, err := ioutil.ReadFile(filePath)
	if err != nil {
		return "", err
	}
	content := string(data)
	return content, nil
}

func folderExists(folder string) bool {
	_, err := os.Stat(folder)
	return err == nil
}

func createFolder(folder string) {
	err := os.MkdirAll(folder, 0755)
	if err != nil {
		panic(err)
	}
}

func fileExists(file string) bool {
	_, err := os.Stat(file)
	return err == nil
}

func downloadFile(url string, outputPath string) error {
	out, err := os.Create(outputPath)
	if err != nil {
		return err
	}
	defer out.Close()

	resp, err := http.Get(url)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("failed to download file, status code: %d", resp.StatusCode)
	}

	_, err = io.Copy(out, resp.Body)
	if err != nil {
		return err
	}

	return nil
}

func clearConsole() {
	switch runtime.GOOS {
	case "windows":
		cmd := exec.Command("cmd", "/c", "cls")
		cmd.Stdout = os.Stdout
		cmd.Run()
	case "linux", "darwin":
		cmd := exec.Command("clear")
		cmd.Stdout = os.Stdout
		cmd.Run()
	default:
		fmt.Println("Unable to clear console. Unsupported operating system.")
	}
}
