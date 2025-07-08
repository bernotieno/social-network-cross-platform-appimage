// Theme management for desktop messenger
class ThemeManager {
    constructor() {
        this.currentTheme = 'light';
        this.isElectron = typeof window !== 'undefined' && window.electronAPI;
        this.init();
    }

    async init() {
        // Load saved theme preference
        await this.loadThemePreference();
        
        // Apply the theme
        this.applyTheme(this.currentTheme);
        
        // Listen for system theme changes
        this.setupSystemThemeListener();
    }

    async loadThemePreference() {
        try {
            if (this.isElectron) {
                // Use Electron's secure storage
                const result = await window.electronAPI.secureStorage.getItem('theme');
                if (result.success && result.data) {
                    this.currentTheme = result.data;
                }
            } else {
                // Fallback to localStorage for web compatibility
                const savedTheme = localStorage.getItem('socialNetworkTheme');
                if (savedTheme) {
                    this.currentTheme = savedTheme;
                }
            }
        } catch (error) {
            console.warn('Failed to load theme preference:', error);
            // Default to system preference
            this.currentTheme = this.getSystemTheme();
        }
    }

    async saveThemePreference(theme) {
        try {
            if (this.isElectron) {
                await window.electronAPI.secureStorage.setItem('theme', theme);
            } else {
                localStorage.setItem('socialNetworkTheme', theme);
            }
        } catch (error) {
            console.warn('Failed to save theme preference:', error);
        }
    }

    getSystemTheme() {
        if (window.matchMedia && window.matchMedia('(prefers-color-scheme: dark)').matches) {
            return 'dark';
        }
        return 'light';
    }

    setupSystemThemeListener() {
        if (window.matchMedia) {
            const mediaQuery = window.matchMedia('(prefers-color-scheme: dark)');
            mediaQuery.addEventListener('change', (e) => {
                // Only auto-switch if user hasn't manually set a preference
                if (!this.hasManualThemePreference()) {
                    const newTheme = e.matches ? 'dark' : 'light';
                    this.setTheme(newTheme, false); // Don't save as manual preference
                }
            });
        }
    }

    async hasManualThemePreference() {
        try {
            if (this.isElectron) {
                const result = await window.electronAPI.secureStorage.getItem('theme');
                return result.success && result.data;
            } else {
                return localStorage.getItem('socialNetworkTheme') !== null;
            }
        } catch (error) {
            return false;
        }
    }

    applyTheme(theme) {
        const root = document.documentElement;
        
        // Remove existing theme classes
        root.classList.remove('theme-light', 'theme-dark');
        
        // Add new theme class
        root.classList.add(`theme-${theme}`);
        
        // Update data attribute for CSS targeting
        root.setAttribute('data-theme', theme);
        
        // Dispatch theme change event
        window.dispatchEvent(new CustomEvent('themeChanged', { 
            detail: { theme, previousTheme: this.currentTheme } 
        }));
        
        this.currentTheme = theme;
    }

    async setTheme(theme, savePreference = true) {
        if (theme !== 'light' && theme !== 'dark') {
            console.warn('Invalid theme:', theme);
            return;
        }

        this.applyTheme(theme);
        
        if (savePreference) {
            await this.saveThemePreference(theme);
        }
    }

    async toggleTheme() {
        const newTheme = this.currentTheme === 'light' ? 'dark' : 'light';
        await this.setTheme(newTheme);
        return newTheme;
    }

    getCurrentTheme() {
        return this.currentTheme;
    }

    isDarkMode() {
        return this.currentTheme === 'dark';
    }

    isLightMode() {
        return this.currentTheme === 'light';
    }
}

// Create global theme manager instance
const themeManager = new ThemeManager();

// Export for use in other modules
if (typeof module !== 'undefined' && module.exports) {
    module.exports = ThemeManager;
} else {
    window.themeManager = themeManager;
}
