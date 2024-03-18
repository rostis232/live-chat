document.addEventListener('htmx:beforeSend', function(event) {
    document.getElementById('text').value = '';
    document.getElementById('name').setAttribute('hidden', 'true');
    });