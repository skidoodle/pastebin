(function () {
  const code = document.getElementById('code-block');
  if (code) {
    const ln = document.getElementById('line-numbers');
    const lineCount = parseInt(code.getAttribute('data-line-count'));

    const nums = [];
    for (let i = 1; i <= lineCount; i++) {
      nums.push('<a href="#L' + i + '" id="L' + i + '">' + i + '</a>');
    }
    ln.innerHTML = nums.join('');

    const rawText = code.textContent;
    const workerCode = `
            self.window = self;
            self.document = {
                readyState: 'complete',
                querySelectorAll: function() { return []; },
                addEventListener: function() {}
            };
            importScripts('${window.location.origin}/static/highlight.min.js');
            onmessage = function(e) {
                try {
                    const result = self.hljs.highlightAuto(e.data);
                    postMessage(result.value);
                } catch (err) {
                    postMessage(e.data);
                }
            }
        `;
    const blob = new Blob([workerCode], { type: 'application/javascript' });
    const workerUrl = URL.createObjectURL(blob);
    const worker = new Worker(workerUrl);

    worker.onmessage = function (e) {
      code.innerHTML = e.data;
      URL.revokeObjectURL(workerUrl);
    };
    worker.postMessage(rawText);

    function updateActiveLine() {
      const active = document.querySelector('.line-numbers a.active');
      if (active) active.classList.remove('active');

      if (window.location.hash) {
        const target = document.getElementById(window.location.hash.substring(1));
        if (target) {
          target.scrollIntoView({ block: 'center', behavior: 'smooth' });
          target.classList.add('active');
        }
      }
    }

    updateActiveLine();
    window.addEventListener('hashchange', updateActiveLine);

    document.addEventListener('keydown', (e) => {
      if ((e.ctrlKey || e.metaKey) && e.key === 'a') {
        if (e.target.tagName !== 'TEXTAREA' && e.target.tagName !== 'INPUT') {
          e.preventDefault();
          const range = document.createRange();
          range.selectNodeContents(code);
          const selection = window.getSelection();
          selection.removeAllRanges();
          selection.addRange(range);
        }
      }
    });
  }

  document.addEventListener('keydown', (e) => {
    if ((e.ctrlKey || e.metaKey) && e.key === 's') {
      const form = document.getElementById('paste-form');
      if (form) {
        e.preventDefault();
        form.submit();
      }
    }
  });
})();
