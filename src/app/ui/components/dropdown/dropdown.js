// Dropdown web component
class Dropdown extends HTMLElement {
    constructor() {
        super();
        this.attachShadow({ mode: 'open' });

        // Styles (copied from .model-selector in styles.css)
        const style = document.createElement('style');
        style.textContent = `
            @import url('global.css');
            @import url('components/dropdown/dropdown.css');
        `;
        this.shadowRoot.appendChild(style);

        this.wrapper = document.createElement('div');
        this.wrapper.classList.add('dropdown-wrapper');
        this.shadowRoot.appendChild(this.wrapper);

        this.select = document.createElement('select');
        this.wrapper.appendChild(this.select);

        this.select.addEventListener('change', (e) => {
            this.dispatchEvent(new Event('change', { bubbles: true }));
        });
    }

    static get observedAttributes() {
        return ['disabled'];
    }

    attributeChangedCallback(name, oldValue, newValue) {
        if (name === 'disabled') {
            this.select.disabled = this.hasAttribute('disabled');
        }
    }

    set options(opts) {
        this.select.innerHTML = '';
        for (const opt of opts) {
            const option = document.createElement('option');
            option.value = opt.value;
            option.textContent = opt.label;
            if (opt.selected) option.selected = true;
            if (opt.disabled) option.disabled = true;
            this.select.appendChild(option);
        }
    }

    get value() {
        return this.select.value;
    }

    set value(val) {
        this.select.value = val;
    }

    focus() {
        this.select.focus();
    }
}

customElements.define('ui-dropdown', Dropdown);