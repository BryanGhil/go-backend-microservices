# E-Commerce Microservices Platform

## Project Overview
This is a backend e-commerce application built with **Golang**, designed around a microservices architecture. It uses an API Gateway built with the **Gin** framework to handle incoming user requests. The gateway communicates with internal domain services—like Order, Payment, and Product using **gRPC** for fast and efficient internal communication.

To make the system easy to understand and test for frontend developers or external clients, the API gateway endpoints are documented using **Swagger**.

To manage the checkout process and keep data consistent across the different services, the system uses **Apache Kafka** as a message broker. This ensures that actions like processing payments and updating stock happen reliably, and helps safely reverse steps if part of an order fails.

The project also includes **Jaeger** for distributed tracing. This tool monitors how requests travel between the different services, making it much easier to debug issues.

## Core Technologies
* **Language:** Golang (Go)
* **API Gateway:** Gin Web Framework
* **API Documentation:** Swagger
* **Internal Communication:** gRPC & Protocol Buffers
* **Message Broker:** Apache Kafka
* **Monitoring & Tracing:** Jaeger