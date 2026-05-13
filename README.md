# Live Message Display System

## Project Overview

โปรเจกต์นี้เป็นเว็บไซต์สำหรับแสดงข้อความและรูปภาพแบบ Real-time โดยแบ่งผู้ใช้งานออกเป็น 2 ฝั่ง

1. Admin
2. Guest

ระบบถูกออกแบบมาเพื่อใช้แสดงข้อความหรือรูปภาพบนหน้าจอหลัก เช่น

* งาน Event
* ร้านอาหาร / ร้านเหล้า
* Live Wall
* จอขึ้นข้อความ
* หน้าจอแสดงกิจกรรม
* Interactive Screen

เมื่อ Guest ส่งข้อมูลเข้ามา ข้อมูลจะไปแสดงบนหน้าจอ Admin ทันทีแบบ Real-time โดยไม่ต้อง Refresh หน้าเว็บ

---

# System Flow

```txt
Guest Send Message
        ↓
Backend API
        ↓
MongoDB Save
        ↓
WebSocket Broadcast
        ↓
Admin Screen Update Realtime
```

---

# User Roles

## 1. Admin

Admin เป็นผู้ดูแลระบบและควบคุมหน้าจอแสดงผล

### Admin Features

* Login ด้วย Username และ Password
* ดูข้อความและรูปที่ Guest ส่งเข้ามา
* แสดงผลบนจอใหญ่ได้
* เปลี่ยนข้อความที่กำลังแสดง
* ย้อนกลับไป Object ก่อนหน้า
* ลบข้อความหรือรูปที่เลือก
* ดูข้อมูลแบบ Real-time

---

## 2. Guest

Guest สามารถส่งรูปภาพและข้อความเข้าสู่ระบบได้

### Guest Features

* อัปโหลดรูปภาพ
* พิมพ์ข้อความ
* จำกัดข้อความไม่เกิน 30 ตัวอักษร
* กดส่งข้อมูล
* มี Cooldown 30 วินาทีต่อการส่ง 1 ครั้ง

---

# Core Features

# 1. Authentication System

ระบบ Login สำหรับ Admin

## Features

* Username Login
* Password Login
* JWT Authentication
* Protected Routes
* Session Validation

## Tasks

* [ ] สร้าง Login API
* [ ] Hash Password
* [ ] JWT Middleware
* [ ] Admin Session
* [ ] Logout System

---

# 2. Image Storage System

ระบบจัดเก็บรูปภาพของ Guest ลงในเซิร์ฟเวอร์

## Flow

```txt
Guest Upload Image
        ↓
Backend Receive File
        ↓
Check img Folder
        ↓
If Not Exists → Create Folder
        ↓
Generate ID
        ↓
Save Image
        ↓
Save Path To Database
        ↓
Frontend Load Image By Path
```

---

## Image Folder Structure

```txt
backend/
 ├─ img/
 │   ├─ 68231abc.png
 │   ├─ 68231def.jpg
 │
 └─ main.go
```

---

## Image Path Mapping

เมื่อ Guest ส่งรูปเข้ามา ระบบจะ:

1. Generate Object ID
2. ตั้งชื่อไฟล์ตาม ID
3. Save รูปลงโฟลเดอร์ `img`
4. Save Path ลง MongoDB

ตัวอย่างข้อมูลใน Database

```json
{
  "_id": "68231abc",
  "message": "hello",
  "imagePath": "img/68231abc.png"
}
```

---

## Frontend Rendering

Frontend จะดึงข้อมูลจาก Database แล้ว map path รูปโดยตรง

ตัวอย่าง

```tsx
<img src={`http://localhost:3000/${item.imagePath}`} />
```

---

## Backend Requirements

### สิ่งที่ต้องทำ

* [ ] เช็คโฟลเดอร์ img
* [ ] ถ้าไม่มีให้สร้างใหม่อัตโนมัติ
* [ ] Save รูปลง img
* [ ] Generate File Name ด้วย ID
* [ ] Save imagePath ลง Database
* [ ] เปิด Static Route สำหรับรูป

---

# 2. Guest Upload System

ระบบส่งข้อความและรูปภาพ

## Features

* Upload รูปภาพ
* ส่งข้อความ
* จำกัดข้อความ 30 ตัวอักษร
* Cooldown 30 วินาที
* Validate File Type

## Tasks

* [ ] Create Upload Form
* [ ] Validate Text Length
* [ ] Upload Image API
* [ ] Save File
* [ ] Save Message Database
* [ ] Create Cooldown Logic

---

# 3. Real-time System

ระบบอัปเดตหน้าจอทันที

## Features

* Live Update
* Realtime Feed
* WebSocket Broadcast
* Auto Display

## Tasks

* [ ] Setup WebSocket
* [ ] Store Client Connections
* [ ] Broadcast New Message
* [ ] Handle Disconnect
* [ ] Auto Reconnect

---

# 4. Admin Display Screen

หน้าจอแสดงผลหลักสำหรับ Admin

## Features

* แสดงรูปภาพเต็มจอ
* แสดงข้อความ
* เปลี่ยน Object ปัจจุบัน
* ปุ่มย้อนกลับ Object ก่อนหน้า
* ลบ Object
* แสดงผลแบบ Live

## Tasks

* [ ] Create Display Page
* [ ] Fullscreen Mode
* [ ] Previous Object Button
* [ ] Delete Object Button
* [ ] Active Object State
* [ ] Live Rendering

---

# 5. Queue / Object System

ระบบจัดการ Object ที่ถูกส่งเข้ามา

## Object Structure

```json
{
  "id": "string",
  "message": "string",
  "imageUrl": "string",
  "createdAt": "date"
}
```

## Features

* เก็บ Object ทั้งหมด
* เรียงตามเวลา
* เปลี่ยน Object ที่แสดง
* ลบ Object
* ย้อนกลับ Object ก่อนหน้า

---

# Database Design

## Collection: messages

```json
{
  "_id": "ObjectId",
  "message": "string",
  "imageUrl": "string",
  "createdAt": "date"
}
```

---

# API Design

# Admin APIs

## Login

```http
POST /admin/login
```

---

# Guest APIs

## Create Message

```http
POST /messages
```

## Get All Messages

```http
GET /messages
```

## Delete Message

```http
DELETE /messages/:id
```

---

# WebSocket Events

## Connection

```txt
/ws
```

## New Message Event

```txt
new-message
```

## Delete Message Event

```txt
delete-message
```

---

# Frontend State Design

# Guest State

## Upload State

```ts
const [message, setMessage] = useState("")
const [image, setImage] = useState<File | null>(null)
```

---

## Cooldown State

```ts
const [cooldown, setCooldown] = useState(30)
const [canSend, setCanSend] = useState(true)
```

---

## Loading State

```ts
const [loading, setLoading] = useState(false)
```

---

# Admin State

## Message List State

```ts
const [messages, setMessages] = useState([])
```

---

## Current Object State

```ts
const [currentIndex, setCurrentIndex] = useState(0)
```

---

## Selected Object State

```ts
const [selectedObject, setSelectedObject] = useState(null)
```

---

## WebSocket State

```ts
const [connected, setConnected] = useState(false)
```

---

# UI Flow

# First Page

เมื่อเข้าเว็บจะมี 2 ปุ่ม

```txt
[ ADMIN ]
[ GUEST ]
```

---

# Admin Flow

```txt
Enter Admin
     ↓
Login Page
     ↓
Dashboard
     ↓
Display Screen
```

---

# Guest Flow

```txt
Enter Guest
     ↓
Upload Form
     ↓
Select Image
     ↓
Enter Message
     ↓
Send
     ↓
Cooldown 30 Seconds
```

---

# Project Structure

```txt
project/
 ├─ frontend/
 │   ├─ app/
 │   │   ├─ page.tsx
 │   │   ├─ admin/
 │   │   ├─ guest/
 │   │   ├─ login/
 │   │   └─ components/
 │   │
 │   └─ socket/
 │
 ├─ backend/
 │   ├─ handlers/
 │   ├─ services/
 │   ├─ repository/
 │   ├─ websocket/
 │   ├─ middleware/
 │   └─ models/
 │
 ├─ uploads/
 │
 └─ README.md
```

---

# Backend Architecture

## Handler Layer

รับ Request / Response

## Service Layer

Business Logic

## Repository Layer

เชื่อมต่อ Database

## WebSocket Layer

Realtime Broadcast

---

# Development Phases

# Phase 1 — Backend Setup

## Tasks

* [ ] Setup Fiber
* [ ] Connect MongoDB
* [ ] Create Message APIs
* [ ] Create Login APIs
* [ ] Upload File System

---

# Phase 2 — Frontend Setup

## Tasks

* [ ] Create Home Page
* [ ] Create Admin Page
* [ ] Create Guest Page
* [ ] Create Login Page
* [ ] Create Upload Form

---

# Phase 3 — Realtime System

## Tasks

* [ ] Setup WebSocket
* [ ] Broadcast Message
* [ ] Receive Live Message
* [ ] Update Screen Instantly

---

# Phase 4 — Admin Controls

## Tasks

* [ ] Previous Object Button
* [ ] Delete Button
* [ ] Fullscreen Display
* [ ] Queue Navigation

---

# Phase 5 — Production

## Tasks

* [ ] Docker
* [ ] VPS Deploy
* [ ] HTTPS SSL
* [ ] Domain
* [ ] Nginx Reverse Proxy

---

# Future Features

## Possible Upgrades

* Live Animation
* Auto Slide
* Sound Effect
* Online User Count
* Reaction System
* QR Upload
* Video Upload
* Multi Screen Support
* Admin Analytics

---

# Final Goal

สร้างระบบ Live Display Platform ที่สามารถ:

* รับข้อความและรูปภาพจาก Guest
* แสดงผลบนจอ Admin แบบ Real-time
* ควบคุมการแสดงผลผ่าน Admin
* รองรับการใช้งานบนจอใหญ่
* รองรับผู้ใช้หลายคนพร้อมกัน
* ขยายระบบต่อยอดได้ในอนาคต
