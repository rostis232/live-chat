var errorCount = 0;

        document.addEventListener('htmx:wsError', function(event) {
            errorCount++;
            if (errorCount >= 5) {
                var error = event.detail.error;
                var errorMessageElement = document.getElementById('error-message');
                errorMessageElement.textContent = 'Помилка чату: зачекайте або оновіть сторінку (' + error.message + ')';
                errorMessageElement.style.display = 'block';
            }
        });

        document.addEventListener('htmx:wsOpen', function(event) {
            errorCount = 0;
            var errorMessageElement = document.getElementById('error-message');
            errorMessageElement.style.display = 'none';
        });