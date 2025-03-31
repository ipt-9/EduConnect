import { Component, OnInit, AfterViewInit } from '@angular/core';
import { CommonModule } from '@angular/common';

@Component({
  selector: 'app-coding-space',
  standalone: true,
  imports: [CommonModule],
  templateUrl: './coding-space.component.html',
  styleUrls: ['./coding-space.component.scss']
})
export class CodingSpaceComponent implements OnInit, AfterViewInit {
  task: any;
  isLoadingCode: boolean = false;
  submitMessage: string | null = null;
  submitSuccess: boolean | null = null;
  elapsedTime: number = 0;
  timerInterval: any = null;
  formattedTime: string = '00:00';

  ngOnInit(): void {
    const rawTask = localStorage.getItem('activeTask');
    if (!rawTask) {
      console.warn("⚠️ Kein Task in localStorage gefunden.");
      return;
    }

    const storedTask = JSON.parse(rawTask);
    const courseId = 1; // ggf. dynamisch machen
    const token = localStorage.getItem('token');

    if (!token) {
      console.error("⛔️ Kein Token im LocalStorage gefunden.");
      return;
    }

    const headers = new Headers({
      'Authorization': `Bearer ${token}`,
      'Content-Type': 'application/json'
    });

    fetch(`http://localhost:8080/courses/${courseId}/tasks`, { headers })
      .then(res => {
        console.log("📡 Tasks geladen, Status:", res.status);
        return res.json();
      })
      .then((allTasks) => {
        const updatedTask = allTasks.find((t: any) => t.id === storedTask.id);

        if (!updatedTask) {
          console.warn("❌ Aufgabe nicht mehr vorhanden auf Server.");
          return;
        }

        this.task = updatedTask;

        const storedElapsed = localStorage.getItem(`elapsedTime_task_${this.task.id}`);
        this.elapsedTime = storedElapsed ? parseInt(storedElapsed) : 0;
        this.formattedTime = this.formatTime(this.elapsedTime);

        if (!this.task.completed) {
          this.startTimerForTask(this.task.id);
        }

        this.isLoadingCode = true;

        // FETCH: submitted-code vom Server holen mit Debug
        return fetch(`http://localhost:8080/tasks/${this.task.id}/submitted-code`, { headers });
      })
      .then(async res => {
        if (!res) return;
        console.log("📡 submitted-code Status:", res.status);

        const rawText = await res.text();
        console.log("📩 Antwort (Text):", rawText);

        let data;
        try {
          data = JSON.parse(rawText);
        } catch (e) {
          console.error("❌ JSON Parse-Fehler bei submitted-code:", e);
          return;
        }

        if (data?.code !== undefined) {
          console.log("✅ Code erhalten:", data.code);
          this.task.submitted_code = data.code;
        } else {
          console.warn("⚠️ Kein Code-Feld in Antwort enthalten.");
        }
      })
      .catch(err => {
        console.error("❌ Fehler beim Laden von Aufgaben oder Code:", err);
      })
      .finally(() => {
        this.isLoadingCode = false;
      });
  }



  startTimerForTask(taskId: number): void {
    const storedElapsed = localStorage.getItem(`elapsedTime_task_${taskId}`);
    this.elapsedTime = storedElapsed ? parseInt(storedElapsed) : 0;

    const startTimestamp = Date.now();
    localStorage.setItem(`startTime_task_${taskId}`, String(startTimestamp));

    this.timerInterval = setInterval(() => {
      const newElapsed = Math.floor((Date.now() - startTimestamp) / 1000) + this.elapsedTime;
      this.formattedTime = this.formatTime(newElapsed);
      localStorage.setItem(`elapsedTime_task_${taskId}`, String(newElapsed));
    }, 1000);
  }
  stopTimer(taskId: number): void {
    if (this.timerInterval) {
      clearInterval(this.timerInterval);
      this.timerInterval = null;
    }

    const startTimeStr = localStorage.getItem(`startTime_task_${taskId}`);
    if (startTimeStr) {
      const startTime = parseInt(startTimeStr);
      const additionalTime = Math.floor((Date.now() - startTime) / 1000);
      this.elapsedTime += additionalTime;
      localStorage.setItem(`elapsedTime_task_${taskId}`, String(this.elapsedTime));
    }
  }


  ngAfterViewInit(): void {
    const interval = setInterval(() => {
      const codeText = document.getElementById('codeText');
      const runBtn = document.getElementById('runBtn');
      const ghost = document.getElementById('ghost');
      const outputBox = document.querySelector('.output');
      const lineNumbers = document.getElementById('lineNumbers');

      const everythingReady = codeText && runBtn && ghost && outputBox && lineNumbers && this.task;

      if (everythingReady) {
        clearInterval(interval);
        this.initEditor(); // 👈 deine bisherige Logik ausgelagert
      }
    }, 100);
  }
  initEditor(): void {
    const self = this;
    const codeText = document.getElementById('codeText') as HTMLElement;
    const ghost = document.getElementById('ghost') as HTMLElement;
    const lineNumbers = document.getElementById('lineNumbers') as HTMLElement;
    const runBtn = document.getElementById('runBtn') as HTMLButtonElement;
    const outputBox = document.querySelector('.output') as HTMLElement;

    let suggestion = '';

    if (this.task?.submitted_code) {
      codeText.innerText = this.task.submitted_code;
    } else if (this.task?.starter_code) {
      codeText.innerText = this.task.starter_code;
    }

    function updateLineNumbers() {
      const lines = codeText.innerText.split('\n').length || 1;
      lineNumbers.innerText = Array.from({ length: lines }, (_, i) => i + 1).join('\n');
    }

    function updateSuggestion() {
      const text = codeText.innerText;
      const words = text.trim().split(/\s+/);
      const lastWord = words[words.length - 1] || '';

      if (lastWord === 'pri') {
        suggestion = 'nt()';
        ghost.innerText = suggestion;
      } else {
        suggestion = '';
        ghost.innerText = '';
      }
    }

    codeText.addEventListener('keydown', (e) => {
      if (e.key === 'Tab' && suggestion) {
        e.preventDefault();
        document.execCommand('insertText', false, suggestion);
        ghost.innerText = '';
        suggestion = '';
        updateLineNumbers();
      }
    });

    codeText.addEventListener('input', () => {
      updateSuggestion();
      updateLineNumbers();
    });

    codeText.addEventListener('keyup', updateSuggestion);
    codeText.addEventListener('click', updateSuggestion);

    runBtn.addEventListener('click', async () => {
      const fullCode = codeText.innerText;
      const inputRegex = /input\s*\(\s*["'](.*?)["']\s*\)/;
      const match = fullCode.match(inputRegex);

      if (match) {
        const promptText = match[1];
        outputBox.innerHTML = `<span style="color:#ccc">${promptText} </span><input id="consoleInput" type="text" style="background:transparent;border:none;color:#00ff95;font-family:'Fira Code';font-size:0.9rem;width:100px;" autofocus />`;

        const consoleInput = document.getElementById('consoleInput') as HTMLInputElement;

        consoleInput.addEventListener('keydown', async (e) => {
          if (e.key === 'Enter') {
            const userInput = consoleInput.value;
            const updatedCode = fullCode.replace(inputRegex, `"${userInput}"`);
            outputBox.innerHTML = `⏳ Running...`;

            try {
              const res = await fetch('https://emkc.org/api/v2/piston/execute', {
                method: 'POST',
                headers: { 'Content-Type': 'application/json' },
                body: JSON.stringify({
                  language: 'python3',
                  version: '3.10.0',
                  files: [{ content: updatedCode }]
                })
              });

              const data = await res.json();
              const output = data.run.output || '';
              outputBox.innerText = `${promptText} ${userInput}\n${output}`;
              localStorage.setItem('actualOutput', output.trim());
            } catch (err) {
              outputBox.innerText = '⚠️ Error executing code';
            }
          }
        });
      } else {
        outputBox.innerText = '⏳ Running...';

        try {
          const res = await fetch('https://emkc.org/api/v2/piston/execute', {
            method: 'POST',
            headers: { 'Content-Type': 'application/json' },
            body: JSON.stringify({
              language: 'python3',
              version: '3.10.0',
              files: [{ content: fullCode }]
            })
          });

          const data = await res.json();
          const output = data.run.output || 'No output';
          outputBox.innerText = output;
          localStorage.setItem('actualOutput', output.trim());
        } catch (err) {
          outputBox.innerText = '⚠️ Error executing code';
        }
      }
    });

    updateLineNumbers();
    updateSuggestion();
  }


  async executeCode(code: string): Promise<string> {
    try {
      const response = await fetch('https://emkc.org/api/v2/piston/execute', {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({
          language: 'python3',
          version: '3.10.0',
          files: [{ content: code }]
        })
      });

      const data = await response.json();
      const output = data?.run?.output || '';
      localStorage.setItem('actualOutput', output.trim()); // ⚠️ vollständiger Output speichern
      return output;
    } catch (error) {
      console.error('❌ Fehler beim Ausführen des Codes:', error);
      return '⚠️ Fehler bei der Code-Ausführung';
    }
  }

  async submitSolution(): Promise<void> {
    const codeText = document.getElementById('codeText') as HTMLElement;
    const token = localStorage.getItem('token');

    if (!this.task || !codeText || !token) {
      console.error("⛔️ Fehlende Daten beim Submit.");
      return;
    }

    let actualOutput = localStorage.getItem('actualOutput');

    if (!actualOutput) {
      actualOutput = await this.executeCode(codeText.innerText);
    }

    const finalElapsedTime = parseInt(localStorage.getItem(`elapsedTime_task_${this.task.id}`) || '0');
    const executionTimeMs = finalElapsedTime * 1000;

    const body = {
      task_id: this.task.id,
      code: codeText.innerText,
      output: actualOutput,
      execution_time_ms: executionTimeMs,
      used_hint: false
    };

    fetch('http://localhost:8080/submit', {
      method: 'POST',
      headers: {
        'Authorization': `Bearer ${token}`,
        'Content-Type': 'application/json'
      },
      body: JSON.stringify(body)
    })
      .then(res => res.json())
        .then(data => {
          if (data.success) {
            this.stopTimer(this.task.id); // ✅ endgültig stoppen
          }

          this.submitSuccess = data.success;
          this.submitMessage = data.success
            ? "✅ Aufgabe erfolgreich eingereicht!"
            : "❌ Die Lösung war leider falsch.";

          setTimeout(() => {
            this.submitMessage = null;
            this.submitSuccess = null;
          }, 5000);
        })

      .catch(err => {
        console.error("❌ Fehler beim Submit:", err);
        this.submitSuccess = false;
        this.submitMessage = "⚠️ Fehler beim Einreichen der Lösung.";
      });
  }

  formatTime(seconds: number): string {
    const min = Math.floor(seconds / 60);
    const sec = seconds % 60;
    return `${min.toString().padStart(2, '0')}:${sec.toString().padStart(2, '0')}`;
  }
  ngOnDestroy(): void {
    if (this.task?.id && !this.task?.completed) {
      this.stopTimer(this.task.id);
    }
  }
  goBackToTasks(): void {
    window.location.href = '/taskslist';
  }
  resetTimer(): void {
    if (!this.task || this.task.completed) return;

    const key = `elapsedTime_task_${this.task.id}`;

    // Timer stoppen und Werte zurücksetzen
    this.elapsedTime = 0;
    this.formattedTime = this.formatTime(0);
    localStorage.setItem(key, '0');

    if (this.timerInterval) {
      clearInterval(this.timerInterval);
    }

    this.startTimerForTask(this.task.id); // ⏱ Timer direkt neu starten
  }

}
