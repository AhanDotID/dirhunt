# 🔍 DirHunt

```
  ██████╗ ██╗██████╗ ██╗  ██╗██╗   ██╗███╗   ██╗████████╗
  ██╔══██╗██║██╔══██╗██║  ██║██║   ██║████╗  ██║╚══██╔══╝
  ██║  ██║██║██████╔╝███████║██║   ██║██╔██╗ ██║   ██║   
  ██║  ██║██║██╔══██╗██╔══██║██║   ██║██║╚██╗██║   ██║   
  ██████╔╝██║██║  ██║██║  ██║╚██████╔╝██║ ╚████║   ██║   
  ╚═════╝ ╚═╝╚═╝  ╚═╝╚═╝  ╚═╝ ╚═════╝ ╚═╝  ╚═══╝   ╚═╝  
```

**Fast Directory & File Bruteforcer** written in Go

> Built by [Ahan Pahlevi | CianjurSec](https://ahandotid.github.io) - for educational & authorized security testing purposes only.

---

## ✨ Features

- ⚡ **Blazing fast** - powered by Go goroutines (concurrent requests)
- 🎨 **Color-coded output** - 200 green, 403 yellow, 3xx cyan, 5xx purple
- 🔌 **Extension fuzzing** - append `.php`, `.html`, `.txt`, etc.
- 🎯 **Status code filtering** - match or ignore specific codes
- 🔀 **Redirect control** - follow or block redirects
- 💾 **Output to file** - save results for later
- 🔒 **TLS skip** - works on self-signed certs
- 🧠 **Custom User-Agent** - blend in with real traffic

---

## 📦 Installation

### From Source (requires Go 1.18+)
```bash
git clone https://github.com/AhanDotID/dirhunt
cd dirhunt
go build -o dirhunt .
./dirhunt -h
```

### Quick Install
```bash
go install github.com/AhanDotID/dirhunt@latest
```

---

## 🚀 Usage

```
Usage: dirhunt -u <URL> -w <wordlist> [options]

Options:
  -u  string    Target URL (e.g. https://example.com)
  -w  string    Path to wordlist file
  -t  int       Number of concurrent threads (default: 50)
  -x  string    Extensions to append (e.g. php,html,txt)
  -mc string    Match HTTP status codes (e.g. 200,301,403)
  -fc string    Filter/ignore status codes (default: 404)
  -o  string    Output file to save results
  -ua string    Custom User-Agent header
  -timeout int  Request timeout in seconds (default: 10)
  -no-follow    Do not follow redirects
```

---

## 📖 Examples

**Basic scan:**
```bash
./dirhunt -u https://target.com -w wordlists/common.txt
```

**With extensions:**
```bash
./dirhunt -u https://target.com -w wordlists/common.txt -x php,html,txt
```

**Custom threads & timeout:**
```bash
./dirhunt -u https://target.com -w wordlists/common.txt -t 100 -timeout 5
```

**Match only specific status codes:**
```bash
./dirhunt -u https://target.com -w wordlists/common.txt -mc 200,301,403
```

**Save results to file:**
```bash
./dirhunt -u https://target.com -w wordlists/common.txt -o results.txt
```

**Full combo:**
```bash
./dirhunt -u https://target.com -w wordlists/common.txt -x php,bak -t 80 -fc 404,400 -o output.txt
```

---

## 🎨 Output Colors

| Color | Meaning |
|-------|---------|
| 🟢 Green | 2xx - Found |
| 🔵 Cyan | 3xx - Redirect |
| 🟡 Yellow | 403 - Forbidden |
| 🔴 Red | 4xx - Client Error |
| 🟣 Purple | 5xx - Server Error |

---

## 📁 Wordlists

Included in `wordlists/`:
- `common.txt` - Common directories, files, API paths, backup files

Recommended external wordlists:
- [SecLists](https://github.com/danielmiessler/SecLists) - `Discovery/Web-Content/`
- [dirsearch wordlist](https://github.com/maurosoria/dirsearch/blob/master/db/dicc.txt)

---

## ⚠️ Disclaimer

DirHunt is intended for **authorized security testing and educational purposes only**. Always get explicit permission before scanning any target. The author is not responsible for any misuse.

---

## 🛠️ Related Tools

- [SubSleuth](https://github.com/AhanDotID/subsleuth) - Subdomain finder

---

Made with ❤️ by [Ahan Pahlevi](https://ahandotid.github.io)
