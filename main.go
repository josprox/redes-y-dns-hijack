package main

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/tls"
	"crypto/x509"
	"crypto/x509/pkix"
	"fmt"
	"log"
	"math/big"
	"net"
	"net/http"
	"os"
	"os/exec"
	"os/signal"
	"strings"
	"syscall"
	"time"
	"unsafe"
)

const (
	hostsFilePath = `C:\Windows\System32\drivers\etc\hosts`
	hostsTagStart = "# >>> ESTRELLA LOCAL DEMO START >>>"
	hostsTagEnd   = "# <<< ESTRELLA LOCAL DEMO END <<<"
	domainName    = "estrella.com"

	// Constantes de la API de consola de Windows
	ENABLE_QUICK_EDIT_MODE = 0x0040
	ENABLE_EXTENDED_FLAGS  = 0x0080
	STD_INPUT_HANDLE       = ^uintptr(9) // -10 en uintptr
)

// Desactiva el modo "QuickEdit" de Windows automáticamente por código
// para que hacer clic en la ventana de la terminal NUNCA congele el servidor.
func disableQuickEdit() {
	kernel32 := syscall.NewLazyDLL("kernel32.dll")
	getStdHandle := kernel32.NewProc("GetStdHandle")
	getConsoleMode := kernel32.NewProc("GetConsoleMode")
	setConsoleMode := kernel32.NewProc("SetConsoleMode")

	hStdin, _, _ := getStdHandle.Call(uintptr(STD_INPUT_HANDLE))
	if hStdin == 0 || hStdin == ^uintptr(0) {
		return
	}

	var mode uint32
	ret, _, _ := getConsoleMode.Call(hStdin, uintptr(unsafe.Pointer(&mode)))
	if ret == 0 {
		return
	}

	// Quitar ENABLE_QUICK_EDIT_MODE y agregar ENABLE_EXTENDED_FLAGS
	newMode := (mode &^ ENABLE_QUICK_EDIT_MODE) | ENABLE_EXTENDED_FLAGS
	setConsoleMode.Call(hStdin, uintptr(newMode))
}

// Agrega estrella.com al archivo hosts de Windows
func addHostsEntry() error {
	removeHostsEntry() // Limpiar por si quedó algo anterior

	content, err := os.ReadFile(hostsFilePath)
	if err != nil {
		return fmt.Errorf("no se pudo leer el archivo hosts: %w", err)
	}

	entry := fmt.Sprintf("\r\n%s\r\n127.0.0.1 %s www.%s\r\n%s\r\n", hostsTagStart, domainName, domainName, hostsTagEnd)
	newContent := string(content) + entry

	if err := os.WriteFile(hostsFilePath, []byte(newContent), 0644); err != nil {
		return fmt.Errorf("no se pudo escribir en hosts (¿Ejecutaste como Administrador?): %w", err)
	}

	// Limpiar caché DNS de Windows
	exec.Command("ipconfig", "/flushdns").Run()
	return nil
}

// Remueve la entrada de estrella.com del archivo hosts
func removeHostsEntry() {
	content, err := os.ReadFile(hostsFilePath)
	if err != nil {
		return
	}

	lines := strings.Split(string(content), "\n")
	var filteredLines []string
	insideBlock := false

	for _, line := range lines {
		trimmed := strings.TrimSpace(line)
		if trimmed == hostsTagStart {
			insideBlock = true
			continue
		}
		if trimmed == hostsTagEnd {
			insideBlock = false
			continue
		}
		if insideBlock {
			continue
		}
		if strings.Contains(line, "estrella.com") && strings.Contains(line, "127.0.0.1") {
			continue
		}
		filteredLines = append(filteredLines, line)
	}

	cleaned := strings.Join(filteredLines, "\n")
	cleaned = strings.TrimRight(cleaned, "\r\n") + "\r\n"
	os.WriteFile(hostsFilePath, []byte(cleaned), 0644)
	exec.Command("ipconfig", "/flushdns").Run()
}

// Genera un certificado TLS autofirmado dinámico para estrella.com
func generateSelfSignedCert(domain string) (tls.Certificate, error) {
	priv, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		return tls.Certificate{}, err
	}

	template := x509.Certificate{
		SerialNumber: big.NewInt(time.Now().UnixNano()),
		Subject: pkix.Name{
			Organization: []string{"Estrella Local Demo"},
			CommonName:   domain,
		},
		NotBefore:             time.Now().Add(-1 * time.Hour),
		NotAfter:              time.Now().Add(365 * 24 * time.Hour),
		KeyUsage:              x509.KeyUsageKeyEncipherment | x509.KeyUsageDigitalSignature,
		ExtKeyUsage:           []x509.ExtKeyUsage{x509.ExtKeyUsageServerAuth},
		BasicConstraintsValid: true,
		DNSNames:              []string{domain, "www." + domain, "localhost"},
		IPAddresses:           []net.IP{net.ParseIP("127.0.0.1")},
	}

	certDER, err := x509.CreateCertificate(rand.Reader, &template, &template, &priv.PublicKey, priv)
	if err != nil {
		return tls.Certificate{}, err
	}

	return tls.Certificate{
		Certificate: [][]byte{certDER},
		PrivateKey:  priv,
	}, nil
}

// Manejador HTTP/HTTPS
func makeHTMLHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		log.Printf("[WEB] Petición recibida: %s %s | Host: %s | TLS: %v", r.Method, r.URL.Path, r.Host, r.TLS != nil)

		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		html := `<!DOCTYPE html>
<html lang="es">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>Estrella Local</title>
    <style>
        body {
            font-family: 'Segoe UI', Tahoma, Geneva, Verdana, sans-serif;
            background: linear-gradient(135deg, #0d1b2a 0%, #1b263b 50%, #415a77 100%);
            min-height: 100vh;
            margin: 0;
            display: flex;
            align-items: center;
            justify-content: center;
            color: #fff;
        }
        .card {
            background: rgba(255, 255, 255, 0.08);
            backdrop-filter: blur(16px);
            padding: 45px 50px;
            border-radius: 24px;
            box-shadow: 0 12px 40px rgba(0, 0, 0, 0.6);
            border: 1px solid rgba(255, 255, 255, 0.18);
            text-align: center;
            max-width: 520px;
        }
        .star {
            font-size: 3.8rem;
            animation: pulse 2s infinite ease-in-out;
        }
        @keyframes pulse {
            0%, 100% { transform: scale(1); }
            50% { transform: scale(1.15); }
        }
        h1 {
            font-size: 2.8rem;
            margin: 10px 0 5px 0;
            color: #fca311;
            letter-spacing: -1px;
        }
        h2 {
            font-size: 1.5rem;
            margin: 0 0 20px 0;
            color: #e0e1dd;
            font-weight: 400;
        }
        .badge {
            display: inline-block;
            background: #2ec4b6;
            color: #011627;
            padding: 6px 16px;
            border-radius: 50px;
            font-size: 0.9rem;
            font-weight: 700;
            margin-bottom: 20px;
            text-transform: uppercase;
            letter-spacing: 0.5px;
        }
        .info {
            background: rgba(0, 0, 0, 0.4);
            padding: 16px 20px;
            border-radius: 12px;
            font-family: Consolas, monospace;
            font-size: 0.92rem;
            text-align: left;
            line-height: 1.7;
            border-left: 4px solid #fca311;
        }
    </style>
</head>
<body>
    <div class="card">
        <div class="star">⭐</div>
        <h1>hola estrella</h1>
        <h2>estoy en local</h2>
        <div class="badge">Resolución Local Activa (Hosts Hijack)</div>
        <div class="info">
            • <b>Host:</b> ` + r.Host + `<br>
            • <b>Protocolo:</b> ` + r.Proto + ` (HTTPS: ` + fmt.Sprintf("%v", r.TLS != nil) + `)<br>
            • <b>Dirección Cliente:</b> ` + r.RemoteAddr + `<br>
            • <b>Destino:</b> Servidor Go en 127.0.0.1<br>
            • <b>Estado:</b> Interceptado antes de salir a ICANN
        </div>
    </div>
</body>
</html>`
		w.Write([]byte(html))
	}
}

func main() {
	// Desactivar QuickEdit por código para que un clic en la consola no pause el servidor
	disableQuickEdit()

	fmt.Println("==================================================================")
	fmt.Println(" ⭐ DEMO DE REDES: DOMINIO LOCAL 'estrella.com' (HOST-ONLY) ⭐")
	fmt.Println("==================================================================")

	// 1. Modificar hosts de Windows automáticamente
	fmt.Println("[*] Registrando 'estrella.com' en C:\\Windows\\System32\\drivers\\etc\\hosts...")
	if err := addHostsEntry(); err != nil {
		fmt.Printf("[ERROR] %v\n", err)
		fmt.Println("👉 Asegúrate de abrir como Administrador.")
		time.Sleep(5 * time.Second)
		return
	}
	fmt.Println("[+] Dominio 'estrella.com' registrado localmente -> 127.0.0.1")

	// Asegurar limpieza al salir
	cleanupDone := false
	cleanup := func() {
		if !cleanupDone {
			cleanupDone = true
			fmt.Println("\n[*] Restaurando archivo hosts de Windows...")
			removeHostsEntry()
			fmt.Println("[+] Archivo hosts restaurado limpio. 'estrella.com' ya no apunta a local.")
		}
	}
	defer cleanup()

	// Capturar Ctrl+C o cierre de ventana
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)
	go func() {
		<-sigChan
		cleanup()
		os.Exit(0)
	}()

	handler := makeHTMLHandler()

	// 2. Servidor HTTP (Puerto 80)
	go func() {
		server80 := &http.Server{
			Addr:    ":80",
			Handler: handler,
		}
		if err := server80.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Printf("[HTTP :80] Error: %v", err)
		}
	}()

	// 3. Servidor HTTPS (Puerto 443 con TLS autofirmado)
	go func() {
		cert, err := generateSelfSignedCert(domainName)
		if err != nil {
			log.Printf("[HTTPS :443] Error TLS: %v", err)
			return
		}

		tlsConfig := &tls.Config{
			Certificates: []tls.Certificate{cert},
		}

		server443 := &http.Server{
			Addr:      ":443",
			Handler:   handler,
			TLSConfig: tlsConfig,
		}

		ln, err := net.Listen("tcp", ":443")
		if err != nil {
			log.Printf("[HTTPS :443] Error en puerto 443: %v", err)
			return
		}
		tlsLn := tls.NewListener(ln, tlsConfig)
		if err := server443.Serve(tlsLn); err != nil && err != http.ErrServerClosed {
			log.Printf("[HTTPS :443] Error: %v", err)
		}
	}()

	fmt.Println("[+] Servidor Web HTTP listo en puerto 80")
	fmt.Println("[+] Servidor Web HTTPS listo en puerto 443")
	fmt.Println("[+] Protección de congelamiento activa (QuickEdit desactivado por código)")
	fmt.Println("------------------------------------------------------------------")
	fmt.Println("🎉 ¡YA ESTÁ ACTIVO! Entra a:")
	fmt.Println("    👉 https://estrella.com")
	fmt.Println("------------------------------------------------------------------")
	fmt.Println("💡 Cuando termines la demostración, presiona [Ctrl + C].")
	fmt.Println("==================================================================")

	select {}
}
