# 🔗 LinkedIn Automation Tool (Educational Project)

> ⚠️ **DISCLAIMER**  
> This project is developed **strictly for educational and research purposes** as part of an academic assignment.  
> Automating LinkedIn actions violates LinkedIn’s Terms of Service.  
> **Do NOT use this tool on real or production accounts.**

---

## 📌 Project Overview

This project is a **LinkedIn Automation Tool** built using **Go (Golang)** and **Rod (Chrome DevTools Protocol)**.  
The primary objective of this assignment is to **demonstrate anti-bot detection and stealth automation techniques** by simulating realistic human behavior.

### The tool focuses on:
- Human-like interaction simulation  
- Anti-bot detection avoidance  
- Secure session handling  
- Ethical automation design (educational use only)

---

## 🧠 Key Features

- Automated LinkedIn login  
- Session persistence using cookies  
- Human-like mouse and keyboard behavior  
- Randomized timing and interaction patterns  
- Browser fingerprint masking  
- Anti-detection stealth techniques  
- Rate-limited and scheduled automation  

---

## 🛡️ Anti-Bot Detection Strategy  
*(Core Assignment Requirement)*

This project implements **8 stealth techniques**, including **all 3 mandatory techniques** specified in the assignment.

---

## ✅ Mandatory Stealth Techniques

### 1️⃣ Human-Like Mouse Movement
- Simulates curved mouse paths instead of straight lines  
- Adds micro-corrections and natural overshoot  
- Avoids robotic click patterns  

### 2️⃣ Randomized Timing Patterns
- Random delays between actions  
- Variable think time before clicks and typing  
- Random interaction intervals to mimic human cognition  

### 3️⃣ Browser Fingerprint Masking
- Custom User-Agent string  
- Disables `navigator.webdriver`  
- Fixed and randomized viewport dimensions  
- Prevents automation fingerprint detection  

---

## ➕ Additional Stealth Techniques Implemented

### 4️⃣ Random Scrolling Behavior
- Variable scroll speed  
- Natural acceleration and deceleration  
- Occasional scroll-back actions  

### 5️⃣ Realistic Typing Simulation
- Randomized keystroke delays  
- Occasional typos with backspace correction  
- Human typing rhythm simulation  

### 6️⃣ Mouse Hovering & Cursor Wandering
- Random hover over page elements  
- Idle cursor movement  
- Simulates user reading or thinking behavior  

### 7️⃣ Activity Scheduling
- Automation runs only during business hours  
- Random breaks between actions  
- Simulates realistic daily usage patterns  

### 8️⃣ Rate Limiting & Throttling
- Limits number of actions per hour/day  
- Cooldown periods between operations  
- Prevents aggressive automation patterns  

---

## 🧩 Project Structure

```text
linkedin-automation/
│
├── main.go
├── go.mod
├── README.md
│
├── config/
│   └── config.go
│
├── auth/
│   └── login.go
│
├── internal/
│   ├── stealth/
│   │   ├── mouse.go
│   │   ├── typing.go
│   │   ├── timing.go
│   │   ├── scroll.go
│   │   ├── fingerprint.go
│   │   ├── schedule.go
│   │   └── rate_limit.go
│   │
│   └── logger/
│       └── logger.go
│
├── data/
│   └── session.json
│
└── .env.example

⚙️ Environment Setup
🔐 Create .env file

Create a .env file in the root directory:

LINKEDIN_EMAIL=your_email@example.com
LINKEDIN_PASSWORD=your_password


❗ Never commit your real credentials to version control.

🚀 How to Run the Project
1️⃣ Install Go:
Make sure Go is installed:
go version

2️⃣ Download Dependencies
go mod tidy

3️⃣ Run the Application
go run main.go

📦 Session Management:
Login session cookies are stored locally
Automatic session reuse to avoid repeated logins
Reduces login frequency and detection risk

🧪 Testing Strategy:

Tested on Chromium via Rod
Manual observation of human-like behavior
Logs used to verify randomized timings and actions

📚 Technologies Used

Go (Golang)
Rod (Chrome DevTools Protocol)
Chromium
dotenv (Environment Variables)

⚠️ Ethical & Legal Notice

This project:
Is intended only for learning and demonstration
Must not be used for real automation
Does not promote misuse of LinkedIn services

👩‍💻 Author

Priya Vora
3rd Year Computer Engineering Student
Academic Project – Automation & Anti-Bot Detection

⭐ Final Notes

This project demonstrates how automation can be made indistinguishable from human behavior using layered stealth techniques.
The focus is on learning browser automation internals and detection avoidance strategies, not real-world misuse.
