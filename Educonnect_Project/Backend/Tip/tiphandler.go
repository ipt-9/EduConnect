package Tip

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/ipt-9/EduConnect/DB"
	"net/http"
	"os"
	"strings"
	"time"
)

func BuildGeminiPrompt(task DB.Task, userCode, expectedOutput, actualOutput string) string {
	return fmt.Sprintf(`Ich arbeite an einer Python-Programmieraufgabe und habe meine Lösung eingereicht. Die Ausgabe entspricht leider nicht der erwarteten.

Kannst du mir bitte **einen hilfreichen Verbesserungstipp geben**, der mich auf die richtige Spur bringt?  
Der Tipp darf ruhig konkret sein und auch den richtigen Funktionsnamen, Format oder Wert enthalten (z. B. diee Ausgabe befehl oder die erwartete Ausgabe),  
soll aber **nicht wie ein Fehlerbericht** klingen – sondern wie ein unterstützender Hinweis.

Der Tipp soll motivierend, konkret und lehrreich formuliert sein.

Hier sind die Aufgabendetails:

Aufgabe: %s  
Beschreibung: %s  
Schwierigkeit: %s  
Thema: %s

Erwartete Ausgabe:
%s

Tatsächliche Ausgabe:
%s

Mein Code:
%s

Bitte gib genau **einen Tipp**, der mir hilft, den Fehler zu erkennen und zu korrigieren.`,
		task.Title,
		task.Description,
		task.Difficulty,
		task.Topic,
		expectedOutput,
		actualOutput,
		userCode,
	)
}

func FetchTipFromGemini(prompt string) (string, error) {
	apiKey := os.Getenv("GEMINI_API_KEY")
	apiURL := "https://generativelanguage.googleapis.com/v1beta/models/gemini-2.0-flash:generateContent?key=" + apiKey

	requestBody := map[string]interface{}{
		"contents": []map[string]interface{}{
			{
				"parts": []map[string]string{
					{"text": prompt},
				},
			},
		},
	}

	jsonBody, _ := json.Marshal(requestBody)
	req, _ := http.NewRequest("POST", apiURL, bytes.NewBuffer(jsonBody))
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	var response map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
		return "", err
	}

	if candidates, ok := response["candidates"].([]interface{}); ok && len(candidates) > 0 {
		if content, ok := candidates[0].(map[string]interface{})["content"].(map[string]interface{}); ok {
			if parts, ok := content["parts"].([]interface{}); ok && len(parts) > 0 {
				if text, ok := parts[0].(map[string]interface{})["text"].(string); ok {
					return text, nil
				}
			}
		}
	}

	return "", errors.New("Kein Text von Gemini erhalten")
}
func DetectErrorType(output string, isCorrect bool) string {
	lower := strings.ToLower(output)

	switch {
	case strings.Contains(lower, "nameerror"):
		return "NameError"
	case strings.Contains(lower, "syntaxerror"):
		return "SyntaxError"
	case strings.Contains(lower, "typeerror"):
		return "TypeError"
	case strings.Contains(lower, "indentationerror"):
		return "IndentationError"
	case strings.Contains(lower, "traceback"):
		return "RuntimeError"
	case !isCorrect:
		return "OutputMismatch"
	default:
		return "None"
	}
}

// ExtractFinalErrorLine extrahiert die letzte Zeile mit "Error" oder "Exception"
// Falls kein Fehler enthalten ist, wird der gesamte Output zurückgegeben
func ExtractFinalErrorLine(output string) string {
	lines := strings.Split(strings.TrimSpace(output), "\n")

	for i := len(lines) - 1; i >= 0; i-- {
		line := strings.TrimSpace(lines[i])
		if strings.Contains(line, "Error") || strings.Contains(line, "Exception") {
			return line
		}
	}

	// Kein spezifischer Fehler → Rückgabe des Original-Outputs
	return strings.TrimSpace(output)
}

// ExtractErrorToken extrahiert den konkreten Fehler-Auslöser aus der Fehlermeldung
func ExtractErrorToken(output string) string {
	lower := strings.ToLower(output)

	if strings.Contains(lower, "nameerror") || strings.Contains(lower, "syntaxerror") {
		start := strings.Index(output, "'")
		end := strings.Index(output[start+1:], "'") + start + 1
		if start != -1 && end != -1 && end > start {
			return output[start+1 : end] // z. B. "prin"
		}
	}

	// Kein spezieller Fehler → z. B. OutputMismatch
	return strings.TrimSpace(output)
}
