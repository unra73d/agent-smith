// Dropdown web component

class Dropdown extends HTMLElement {
    constructor() {
        super();
        this.attachShadow({ mode: 'open' });

        this.wrapper = document.createElement('div');
        this.wrapper.classList.add('dropdown-wrapper');
        this.shadowRoot.appendChild(this.wrapper);

        this.select = document.createElement('select');
        this.wrapper.appendChild(this.select);

        this.select.addEventListener('change', (e) => {
            this.dispatchEvent(new Event('change', { bubbles: true }));
        });
        this._initStyle()
    }

    async _initStyle() {
        this.shadowRoot.adoptedStyleSheets = [
            await loadCSS('global.css'),
            await loadCSS('components/dropdown/dropdown.css')
        ];
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