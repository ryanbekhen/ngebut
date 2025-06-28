// ngebut Static Demo JavaScript
console.log('🚀 ngebut static file server is working!');

document.addEventListener('DOMContentLoaded', function() {
    // Add click animation to links
    const links = document.querySelectorAll('a');
    links.forEach(link => {
        link.addEventListener('click', function(e) {
            this.style.transform = 'scale(0.95)';
            setTimeout(() => {
                this.style.transform = 'scale(1)';
            }, 100);
        });
    });

    // Add some interactive features
    const container = document.querySelector('.container');
    if (container) {
        container.addEventListener('mouseenter', function() {
            this.style.boxShadow = '0 12px 40px rgba(0, 0, 0, 0.15)';
        });
        
        container.addEventListener('mouseleave', function() {
            this.style.boxShadow = '0 8px 32px rgba(0, 0, 0, 0.1)';
        });
    }

    // Show a welcome message
    setTimeout(() => {
        console.log('✨ All static assets loaded successfully!');
    }, 500);
});