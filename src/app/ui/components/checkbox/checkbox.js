class Checkbox extends HTMLElement {
    constructor() {
        super();
        this.attachShadow({ mode: 'open' });

        this.checkbox = document.createElement('input');
        this.checkbox.type = 'checkbox';
        this.checkbox.id = 'cb';
        this.label = document.createElement('label');
        this.label.setAttribute('for', 'cb');
        this.label.textContent = this.getAttribute('label') || '';

        this.shadowRoot.appendChild(this.checkbox);
        this.shadowRoot.appendChild(this.label);

        this.checkbox.addEventListener('change', (e) => {
            this.dispatchEvent(new CustomEvent('change', {
                detail: { checked: this.checkbox.checked }
            }));
        });

        this._initStyle()
    }

    async _initStyle() {
        this.shadowRoot.adoptedStyleSheets = [
            await loadCSS('global.css'),
            await loadCSS('components/checkbox/checkbox.css')
        ];
    }

    static get observedAttributes() {
        return ['checked', 'disabled', 'label'];
    }

    attributeChangedCallback(name, oldValue, newValue) {
        if (name === 'checked') {
            this.checkbox.checked = newValue !== null;
        }
        if (name === 'disabled') {
            this.checkbox.disabled = newValue !== null;
        }
        if (name === 'label') {
            this.label.textContent = newValue || '';
        }
    }

    get checked() {
        return this.checkbox.checked;
    }
    set checked(val) {
        this.checkbox.checked = val;
        if (val) this.setAttribute('checked', '');
        else this.removeAttribute('checked');
    }

    get disabled() {
        return this.checkbox.disabled;
    }
    set disabled(val) {
        this.checkbox.disabled = val;
        if (val) this.setAttribute('disabled', '');
        else this.removeAttribute('disabled');
    }
}

customElements.define('ui-checkbox', Checkbox);