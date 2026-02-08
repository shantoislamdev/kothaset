document.addEventListener('DOMContentLoaded', () => {
  // Initialize Icons
  if (window.lucide) {
    window.lucide.createIcons();
  }

  // Header Scroll Effect
  const header = document.getElementById('main-header');
  window.addEventListener('scroll', () => {
    if (window.scrollY > 10) {
      header.classList.add('scrolled');
    } else {
      header.classList.remove('scrolled');
    }
  });

  // Mobile Menu Toggle
  const mobileToggle = document.getElementById('mobile-toggle');
  const mobileMenu = document.getElementById('mobile-menu');
  const menuIcon = document.getElementById('menu-icon');
  const closeIcon = document.getElementById('close-icon');

  if (mobileToggle && mobileMenu) {
    mobileToggle.addEventListener('click', () => {
      mobileMenu.classList.toggle('hidden');
      header.classList.toggle('mobile-open');
      
      // Toggle icons
      if (mobileMenu.classList.contains('hidden')) {
        menuIcon.classList.remove('hidden');
        closeIcon.classList.add('hidden');
      } else {
        menuIcon.classList.add('hidden');
        closeIcon.classList.remove('hidden');
      }
    });
  }

  // Tabs System
  const tabGroups = document.querySelectorAll('[data-tabs]');
  
  tabGroups.forEach(group => {
    const tabs = group.querySelectorAll('[data-tab-target]');
    const contents = group.querySelectorAll('[data-tab-content]');
    // Also include the default checked one (first one usually doesn't have data-tab-content attr explicitly linking if they are just siblings, 
    // but here we used IDs matching targets)
    
    tabs.forEach(tab => {
      tab.addEventListener('click', () => {
        const targetSelector = tab.getAttribute('data-tab-target');
        const targetContent = group.querySelector(targetSelector);
        
        // Deactivate all
        tabs.forEach(t => t.classList.remove('active'));
        group.querySelectorAll('.install-cmd-box').forEach(c => c.classList.add('hidden'));
        
        // Activate clicked
        tab.classList.add('active');
        if (targetContent) {
            targetContent.classList.remove('hidden');
        }
      });
    });
  });

  // Copy to Clipboard Utility
  window.copyToClipboard = async (text, btnElement) => {
    try {
      await navigator.clipboard.writeText(text);
      
      // Visual feedback
      const originalIcon = btnElement.innerHTML;
      // Use semantic class .text-success instead of Tailwind
      btnElement.innerHTML = '<i data-lucide="check" width="18" height="18" class="text-success"></i>';
      
      // Refresh icons for the new checkmark
      if (window.lucide) {
        window.lucide.createIcons();
      }
      
      setTimeout(() => {
        btnElement.innerHTML = originalIcon;
        // Refresh icons effectively restores the original icon svg
        if (window.lucide) {
           window.lucide.createIcons();
        }
      }, 2000);
    } catch (err) {
      console.error('Failed to copy:', err);
    }
  };

  // Terminal Animation for Feature Card
  const terminal = document.querySelector('.feature-terminal');
  if (terminal) {
      initTerminalAnimation(terminal);
  }

  // Handle "Coming Soon" alerts
  document.addEventListener('click', (e) => {
    const target = e.target.closest('[data-action="alert"]');
    if (target) {
      e.preventDefault();
      alert('This feature is coming soon!');
    }
  });
});

function initTerminalAnimation(element) {
    const steps = [
        { text: 'Run interrupted at 85%...', class: 'opacity-70 mb-2' },
        { text: '$ kothaset generate --resume', class: 'text-terracotta mb-1' },
        { text: 'Resuming from ID #8501', class: 'opacity-70 mt-1' }
    ];
    
    let currentStepIndex = 0;
    
    // Clear initial static content
    element.innerHTML = '';
    
    // Create container for lines
    const contentWrapper = document.createElement('div');
    element.appendChild(contentWrapper);
    
    // Cursor
    const cursor = document.createElement('span');
    cursor.textContent = '▋';
    cursor.className = 'animate-pulse-slow ml-1';
    // cursor needs to move? 
    // Simply appending lines.
    
    const loop = async () => {
        // Reset
        contentWrapper.innerHTML = '';
        currentStepIndex = 0;
        
        // Step 1: Interrupted msg
        await typeLine(contentWrapper, steps[0], 20); // Fast typing or instant?
        await wait(800);
        
        // Step 2: Command
        await typeLine(contentWrapper, steps[1], 50); // Typing effect
        await wait(600);
        
        // Step 3: Resume msg
        await typeLine(contentWrapper, steps[2], 20); 
        
        // Wait before restart
        await wait(4000);
        loop();
    };
    
    loop();
}

function wait(ms) {
    return new Promise(resolve => setTimeout(resolve, ms));
}

async function typeLine(container, step, typeSpeed) {
    const line = document.createElement('div');
    line.className = step.class;
    // Handle classes like 'mb-2' by adding them to line
    // Ensure we handle multiple classes
    step.class.split(' ').forEach(c => line.classList.add(c));
    
    container.appendChild(line);
    
    const text = step.text;
    line.textContent = '';
    
    if (typeSpeed === 0) {
        line.textContent = text;
        return;
    }

    for (let i = 0; i < text.length; i++) {
        line.textContent += text[i];
        await wait(Math.random() * typeSpeed + 20);
    }
}
