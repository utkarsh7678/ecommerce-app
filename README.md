
# 🛍️ E-Commerce Platform

[![License](https://img.shields.io/github/license/yourusername/ecommerce-app)](https://github.com/yourusername/ecommerce-app/blob/main/LICENSE)
[![Issues](https://img.shields.io/github/issues/yourusername/ecommerce-app)](https://github.com/yourusername/ecommerce-app/issues)
[![Stars](https://img.shields.io/github/stars/yourusername/ecommerce-app)](https://github.com/yourusername/ecommerce-app/stargazers)

A **full-stack e-commerce platform** built with **React 19** and **Go (Gin)**, supporting user authentication, product management, shopping cart functionality, and order processing.

---

## ✨ Features

### 🧑‍💼 User Authentication
- 🔐 JWT-based secure login
- 📝 User registration
- 🔒 Protected routes & endpoints

### 🛍️ Product Management
- 🏷️ Browse all available products
- 🔍 Search & view product details

### 🛒 Shopping Cart
- ➕ Add/remove items to cart
- 🔄 Real-time cart updates

### 📦 Order Processing
- 🧾 Create orders
- 📄 View order history & details

---

## 🚀 Tech Stack

### 🔹 Frontend
- **React 19**
- **Material UI (v7)**
- **React Context API** (State Management)
- **React Router v7**
- **React Query** (Data fetching)
- **Formik & Yup** (Form handling & validation)
- **Axios** (HTTP client)
- **React Toastify** (Notifications)

### 🔹 Backend
- **Go 1.21+**
- **Gin Framework**
- **SQLite** (via Modernc driver)
- **JWT** (Authentication)
- **Go Validator** (Input validation)
- **Multipart File Uploads**

---

## 📦 Prerequisites

- **Go** `v1.21+`
- **Node.js** `v18+`
- **SQLite3**
- **Git`**

---

## ⚙️ Installation Guide

### 1️⃣ Clone Repository


### 2️⃣ Backend Setup
```bash
cd backend
go mod download

cp .env   
go run run_migration.go  # Run DB migrations
go run cmd/main.go       # Start the backend server
```

### 3️⃣ Frontend Setup
```bash
cd ../frontend
npm install  # or yarn

cp .env.example .env   # Set VITE_API_URL in .env
npm run dev            # or yarn dev
```

---

## 🌐 Access the Application

- Frontend: http://localhost:3001
- Backend API: http://localhost:8080

---

## 📚 API Endpoints

### 🔐 Authentication
- `POST /api/auth/register` — Register user  
- `POST /api/auth/login` — Login  
- `GET /api/auth/me` — Get current user  

### 📦 Products
- `GET /api/items` — List products  

### 🛒 Cart
- `GET /api/cart` — View user cart  
- `POST /api/cart` — Add item to cart  

### 📄 Orders
- `POST /api/orders` — Place an order  
- `GET /api/orders` — View all orders  
- `GET /api/orders/:id` — Order details  

---

## 📫 Postman Collection

### 🔄 Steps
1. Import the Postman collection from: `docs/ecommerce-api.postman_collection.json`
2. Set environment variable:
   - `base_url` = `http://localhost:8080`

Example: Add to Cart (POST `/api/cart`)
```json
{
  "item_id": 1,
  "quantity": 1
}
```

---

## 🔧 Environment Configuration

### Backend `.env`
```env
PORT=8080
```

### Frontend `.env`
```env
VITE_API_URL=http://localhost:8080
```

---

## 🚀 Production Deployment

### Backend
```bash
go build -o ecommerce-app cmd/main.go
./ecommerce-app
```

### Frontend
```bash
npm run build
# Serve the files from 'dist' using a static file server
```
---

## 📁 Project Structure

```bash
ecommerce-app/
├── backend/                    # Backend server (Go)
│   ├── cmd/
│   │   └── server/
│   │       └── main.go         # Server initialization
│   ├── internal/
│   │   ├── config/             # Configuration files
│   │   ├── handlers/           # HTTP request handlers
│   │   ├── middleware/         # Custom middleware
│   │   └── models/             # Database models
│   ├── db/
│   │   └── migrations/         # DB schema & migration files
│   ├── scripts/                # Utility scripts
│   └── .env                    # Environment variables
│
├── frontend/                   # Frontend application (React)
│   ├── public/                 # Static files (favicon, etc.)
│   └── src/                    # React source files
│       ├── assets/            # Images, fonts, etc.
│       ├── components/        # Reusable UI components
│       │   └── Navbar.jsx     # Navigation bar
│       ├── contexts/          # React Contexts (state management)
│       ├── pages/             # Page-level components
│       │   ├── Login.jsx      # Login page
│       │   ├── ItemsList.jsx  # Product listing
│       │   └── ...            # Other pages
│       ├── App.jsx            # Root component
│       └── main.jsx           # React app entry point
│
├── .gitignore                 # Git ignore rules
└── README.md                 # Project documentation
```

---

## 📄 License

This project is licensed under the **MIT License** – see the [LICENSE](LICENSE) file for details.

---

## 🙏 Acknowledgments

- [Material UI](https://mui.com/)
- [Gin Web Framework](https://gin-gonic.com/)
- [React Query](https://tanstack.com/query)
