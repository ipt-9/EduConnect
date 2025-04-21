  import { Component, OnInit } from '@angular/core';
  import { CommonModule } from '@angular/common';

  @Component({
    selector: 'app-coding-space',
    standalone: true,
    imports: [CommonModule],
    templateUrl: './coding-space.component.html',
    styleUrls: ['./coding-space.component.scss']
  })
  export class CodingSpaceComponent implements OnInit {
    task: any;
    isLoadingCode: boolean = false;
    submitMessage: string | null = null;
    submitSuccess: boolean | null = null;
    elapsedTime: number = 0;
    timerInterval: any = null;
    formattedTime: string = '00:00';
    showAlreadyCompletedPopup: boolean = false;
    showSubmitBlockedPopup: boolean = false;
    runRequested: boolean = false;
    tipTitle = "💡 Hinweis anzeigen";
    tipContent: string | null = null;
    showTipPopup = false;
    showTipDetails = false;
    userTips: string[] = [];
    acknowledgeTip(): void {
      this.showTipPopup = false;
      this.tipTitle = "💡 Hinweis anzeigen";
      this.showTipDetails = false;
    }
    userTipsDetailed: { title: string, text: string, expanded: boolean }[] = [];
    toggleTip(tip: any): void {
      tip.expanded = !tip.expanded;
    }

    confirmTip(): void {
      this.showTipPopup = false;

      // Tipp dauerhaft in der Liste anzeigen (falls nicht doppelt)
      if (this.tipContent && !this.userTips.includes(this.tipContent)) {
        this.userTips.push(this.tipContent);
      }

      // Zur Sicherheit zurücksetzen
      this.tipContent = null;
    }


    ngOnInit(): void {
      const rawTask = localStorage.getItem('activeTask');
      if (!rawTask) {
        console.warn("⚠️ Kein Task in localStorage gefunden.");
        return;
      }

      const storedTask = JSON.parse(rawTask);
      const courseId = localStorage.getItem("activeCourseId");
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
        .then(res => res.json())
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

          fetch(`http://localhost:8080/tasks/${this.task.id}/submitted-code`, { headers })
            .then(async res => {
              if (res.ok) {
                const data = await res.json();
                if (data?.code !== undefined) {
                  this.task.submitted_code = data.code;
                }
              } else if (res.status === 404) {
                this.task.submitted_code = this.task.starter_code;
              }
              this.waitForEditorAndInit();
            })
            .catch(err => {
              console.error("❌ Fehler beim Laden des Codes:", err);
            })
            .finally(() => {
              this.isLoadingCode = false;
            });

          // 💡 Tipps laden
          fetch(`http://localhost:8080/tasks/${this.task.id}/tips`, { headers })
            .then(res => res.json())
            .then((tips) => {
              if (Array.isArray(tips) && tips.length > 0) {
                this.userTipsDetailed = tips.map((tip: any, index: number) => ({
                  title: `Tipp ${index + 1}`,
                  text: tip.text,
                  expanded: false
                }));
              } else {
                this.userTipsDetailed = [];
              }
            })
            .catch(err => {
              console.error("❌ Fehler beim Laden der Tipps:", err);
              this.userTipsDetailed = [];
            });

        })
        .catch(err => {
          console.error("❌ Fehler beim Laden der Aufgaben:", err);
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

    initEditor(): void {
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

      runBtn.addEventListener('click', () => {
        if (this.task?.completed) {
          this.showAlreadyCompletedPopup = true;
        } else {
          this.runCodeNormally();
        }
      });

      updateLineNumbers();
      updateSuggestion();
    }

    runCodeNormally(): void {
      const codeText = document.getElementById('codeText') as HTMLElement;
      const outputBox = document.querySelector('.output') as HTMLElement;
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

        fetch('https://emkc.org/api/v2/piston/execute', {
          method: 'POST',
          headers: { 'Content-Type': 'application/json' },
          body: JSON.stringify({
            language: 'python3',
            version: '3.10.0',
            files: [{ content: fullCode }]
          })
        })
          .then(res => res.json())
          .then(data => {
            const output = data.run.output || 'No output';
            outputBox.innerText = output;
            localStorage.setItem('actualOutput', output.trim());
          })
          .catch(() => {
            outputBox.innerText = '⚠️ Error executing code';
          });
      }
    }

    closePopup(): void {
      this.showAlreadyCompletedPopup = false;
      this.showSubmitBlockedPopup = false;
    }

    runAnyway(): void {
      this.closePopup();
      this.runCodeNormally();
    }

    async submitSolution(): Promise<void> {
      if (this.task?.completed) {
        this.showSubmitBlockedPopup = true;
        return;
      }

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
            this.stopTimer(this.task.id);
          }

          this.submitSuccess = data.success;
          this.submitMessage = data.success
            ? "✅ Aufgabe erfolgreich eingereicht!"
            : "❌ Die Lösung war leider falsch.";

          // 🧠 Zeige generierten Tipp in Popup
          if (!data.success && data.tip) {
            this.tipContent = data.tip;
            this.showTipPopup = true;
          }

          // ✅ Code neu laden, wenn Aufgabe erfolgreich war
          if (data.success) {
            fetch(`http://localhost:8080/tasks/${this.task.id}/submitted-code`, {
              headers: {
                'Authorization': `Bearer ${token}`,
                'Content-Type': 'application/json'
              }
            })
              .then(res => res.json())
              .then(res => {
                if (res.code) {
                  this.task.submitted_code = res.code;
                  this.waitForEditorAndInit();
                }
              })
              .catch(err => {
                console.error("❌ Fehler beim Neuladen des Codes nach Submit:", err);
              });
          }

          // 💡 Tipps automatisch neu laden nach Abgabe (bei Fehler)
          if (!data.success) {
            fetch(`http://localhost:8080/tasks/${this.task.id}/tips`, {
              headers: {
                'Authorization': `Bearer ${token}`,
                'Content-Type': 'application/json'
              }
            })
              .then(res => res.json())
              .then((tips) => {
                if (Array.isArray(tips)) {
                  this.userTipsDetailed = tips.map((tip: any, index: number) => ({
                    title: `Tipp ${index + 1}`,
                    text: tip.text,
                    expanded: false
                  }));
                }
              })
              .catch(err => {
                console.error("❌ Fehler beim Nachladen der Tipps:", err);
              });
          }

          // ❌ Message zurücksetzen nach ein paar Sekunden
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
        localStorage.setItem('actualOutput', output.trim());
        return output;
      } catch (error) {
        console.error('❌ Fehler beim Ausführen des Codes:', error);
        return '⚠️ Fehler bei der Code-Ausführung';
      }
    }

    waitForEditorAndInit(): void {
      const interval = setInterval(() => {
        const codeText = document.getElementById('codeText');
        const runBtn = document.getElementById('runBtn');
        const ghost = document.getElementById('ghost');
        const outputBox = document.querySelector('.output');
        const lineNumbers = document.getElementById('lineNumbers');

        const everythingReady = codeText && runBtn && ghost && outputBox && lineNumbers && this.task;

        if (everythingReady) {
          clearInterval(interval);
          this.initEditor();
        }
      }, 100);
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

      this.elapsedTime = 0;
      this.formattedTime = this.formatTime(0);
      localStorage.setItem(key, '0');

      if (this.timerInterval) {
        clearInterval(this.timerInterval);
      }

      this.startTimerForTask(this.task.id);
    }
  }
