# 🌟 Estrella Local Demo (Redes y DNS Hijack)

Proyecto práctico para clases y demostraciones de redes, resolución DNS, certificados TLS y arquitectura cliente-servidor en **Go**.

Permite interceptar el dominio `estrella.com` a nivel local (evitando la resolución externa de la ICANN) y servir una página web propia tanto por **HTTP (80)** como por **HTTPS (443)** con certificado TLS generado dinámicamente en memoria.

---

## 🚀 Características

- **Host-Only / Local Hijack**: Modifica temporalmente el archivo `hosts` del sistema y limpia la caché DNS (`ipconfig /flushdns`).
- **Autolimpiante**: Al presionar `Ctrl + C`, restaura automáticamente el archivo `hosts` original.
- **Soporte HTTP y HTTPS**: Servidores concurrentes en los puertos 80 y 443 con certificado TLS autofirmado en tiempo de ejecución.
- **Protección contra congelamiento**: Desactiva automáticamente el modo `QuickEdit` de la consola de Windows para evitar pausas accidentales por clics del usuario.

---

## 📋 Requisitos

- Windows 10/11
- [Go](https://go.dev/dl/) 1.20 o superior
- Permisos de Administrador (necesarios para abrir puertos privilegiados 80/443 y modificar `hosts`)

---

## 🛠️ Ejecución

1. Clona el repositorio:
   ```bash
   git clone https://github.com/josprox/redes-y-dns-hijack.git
   cd redes-y-dns-hijack
   ```

2. Abre una terminal de **PowerShell** o **Símbolo del sistema** como **Administrador**.

3. Compila y ejecuta:
   ```powershell
   go run main.go
   ```
   *(o compila el ejecutable con `go build -o estrella.exe main.go` y ejecuta `.\estrella.exe`)*

4. Abre tu navegador y navega a:
   ```text
   https://estrella.com
   ```
   *(Acepta la advertencia del certificado autofirmado en "Configuración avanzada" -> "Continuar").*

5. Para detener la demo y restaurar la configuración:
   Presiona `Ctrl + C` en la terminal.
