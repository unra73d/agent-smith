// EditDialog web component
class EditDialog extends HTMLElement {
    constructor() {
        super();
        this.attachShadow({ mode: 'open' });

        // Styles (reuse mcp style approach)
        const style = document.createElement('style');
        style.textContent = `
            @import url('global.css');
            @import url('components/edit-dialog/edit-dialog.css');
        `;
        this.shadowRoot.appendChild(style);

        // Dialog container
        this.dialog = document.createElement('div');
        this.dialog.className = 'dialog';
        this.shadowRoot.appendChild(this.dialog);

        // Prevent click outside from closing by default
        this.addEventListener('mousedown', e => {
            if (e.target !== this) this._cancel();
        });
    }

    // Show dialog with fields, initial values, and custom buttons
    open({ title = 'Edit', fields = [], values = {}, buttons = [] }) {
        this._fields = fields;
        this._values = { ...values };
        this._customButtons = buttons;

        this.dialog.innerHTML = '';

        // Title
        const titleEl = document.createElement('div');
        titleEl.textContent = title;
        titleEl.style.cssText = 'font-size:1.2em;font-weight:600;margin-bottom:8px;color:#e3e3e3;';
        this.dialog.appendChild(titleEl);

        // Fields
        const fieldsEl = document.createElement('div');
        fieldsEl.className = 'fields';
        this._inputEls = {};
        for (const field of fields) {
            const row = document.createElement('div');
            row.className = field.type === 'checkbox' ? 'field-row checkbox-row' : 'field-row';

            // Label
            const label = document.createElement('label');
            label.textContent = field.label + (field.required ? ' *' : '');
            label.htmlFor = field.name;

            // Input
            let input;
            if (field.type === 'select' && Array.isArray(field.options)) {
                input = document.createElement('ui-dropdown');
                input.id = field.name;
                input.options = field.options.map(opt => ({
                    value: opt.value,
                    label: opt.label,
                    selected: (this._values[field.name] ?? field.default ?? '') === opt.value
                }));
                input.value = this._values[field.name] ?? field.default ?? '';
            } else if (field.type === 'checkbox') {
                input = document.createElement('ui-checkbox');
                input.id = field.name;
                input.checked = !!(this._values[field.name] ?? field.default);
                row.appendChild(input);
                row.appendChild(label);
                this._inputEls[field.name] = input;
                fieldsEl.appendChild(row);
                continue;
            } else {
                input = document.createElement('input');
                input.type = field.type === 'password' ? 'password' : 'text';
                input.id = field.name;
                input.value = this._values[field.name] ?? field.default ?? '';
            }
            row.appendChild(label);
            row.appendChild(input);
            this._inputEls[field.name] = input;
            fieldsEl.appendChild(row);
        }
        this.dialog.appendChild(fieldsEl);

        // Buttons
        const btns = document.createElement('div');
        btns.className = 'dialog-buttons';

        // Custom buttons
        for (const btn of buttons) {
            const b = document.createElement('button');
            b.textContent = btn.name;
            b.onclick = async () => {
                const values = this._collectValues();
                await btn.onClick?.(values, this);
            };
            btns.appendChild(b);
        }

        // OK/Cancel
        const okBtn = document.createElement('button');
        okBtn.textContent = 'OK';
        okBtn.onclick = () => this._ok();
        btns.appendChild(okBtn);

        const cancelBtn = document.createElement('button');
        cancelBtn.textContent = 'Cancel';
        cancelBtn.onclick = () => this._cancel();
        btns.appendChild(cancelBtn);

        this.dialog.appendChild(btns);

        // Focus first input
        setTimeout(() => {
            const first = Object.values(this._inputEls)[0];
            if (first && first.focus) first.focus();
        }, 0);

        return new Promise((resolve, reject) => {
            this._resolve = resolve;
            this._reject = reject;
        });
    }

    _collectValues() {
        const result = {};
        for (const field of this._fields) {
            const el = this._inputEls[field.name];
            if (!el) continue;
            if (field.type === 'checkbox') {
                result[field.name] = el.checked;
            } else {
                result[field.name] = el.value;
            }
        }
        return result;
    }

    _ok() {
        // Validate required
        for (const field of this._fields) {
            if (field.required) {
                const val = this._inputEls[field.name];
                if (field.type === 'checkbox') continue;
                if (!val || !val.value) {
                    this._showError(`Field "${field.label}" is required`);
                    return;
                }
            }
        }
        this._resolve(this._collectValues());
        this.remove();
    }

    _cancel() {
        this._resolve(null);
        this.remove();
    }

    _showError(msg) {
        // Simple error display
        let err = this.shadowRoot.querySelector('.dialog-error');
        if (!err) {
            err = document.createElement('div');
            err.className = 'dialog-error';
            err.style.cssText = 'color:#ff6b6b;font-size:0.95em;margin-bottom:4px;';
            this.dialog.insertBefore(err, this.dialog.children[1]);
        }
        err.textContent = msg;
    }
}

customElements.define('edit-dialog', EditDialog);

// Helper function to show dialog and return promise
window.showEditDialog = function ({ title, fields, values, buttons }) {
    return new Promise(resolve => {
        const dlg = document.createElement('edit-dialog');
        document.body.appendChild(dlg);
        dlg.open({ title, fields, values, buttons }).then(resolve);
    });
};