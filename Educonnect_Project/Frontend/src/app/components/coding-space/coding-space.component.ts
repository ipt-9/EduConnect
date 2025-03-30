import { Component, AfterViewInit } from '@angular/core';


@Component({
  selector: 'app-coding-space', // this matches your HTML tag
  templateUrl: './coding-space.component.html',
  styleUrls: ['./coding-space.component.scss']
})

export class CodingSpaceComponent implements AfterViewInit {

  ngAfterViewInit(): void {
    const codeText = document.getElementById('codeText') as HTMLElement;
    const ghost = document.getElementById('ghost') as HTMLElement;
    const lineNumbers = document.getElementById('lineNumbers') as HTMLElement;
    const runBtn = document.getElementById('runBtn') as HTMLButtonElement;
    const outputBox = document.querySelector('.output') as HTMLElement;

    let suggestion = '';

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
              outputBox.innerText = `${promptText} ${userInput}\n${data.run.output || ''}`;
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
          outputBox.innerText = data.run.output || 'No output';
        } catch (err) {
          outputBox.innerText = '⚠️ Error executing code';
        }
      }
    });

    updateLineNumbers();
    updateSuggestion();
  }
}
