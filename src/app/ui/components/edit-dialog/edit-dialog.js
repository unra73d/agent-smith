class EditDialog extends HTMLElement {
    constructor() {
        super();
        this.attachShadow({ mode: 'open' });

        this.dialog = document.createElement('div');
        this.dialog.className = 'dialog';
        this.shadowRoot.appendChild(this.dialog);

        // Error/Status area
        this.statusEl = document.createElement('div');
        this.statusEl.className = 'dialog-status';
        this.statusEl.style.cssText = 'min-height:1.2em;margin-bottom:6px;font-size:0.98em;';
        this.dialog.appendChild(this.statusEl);

        this.addEventListener('mousedown', e => {
            if (e.target !== this) this._cancel();
        });
        this._initStyle()
    }

    async _initStyle() {
        this.shadowRoot.adoptedStyleSheets = [
            await loadCSS('global.css'),
            await loadCSS('components/edit-dialog/edit-dialog.css')
        ];
    }

    open({ title = 'Edit', fields = [], values = {}, buttons = [], validate = null, onClose = null }) {
        this._fields = fields;
        this._values = { ...values };
        this._customButtons = buttons;
        this._validate = validate;
        this._onClose = onClose;

        this.dialog.innerHTML = '';
        this.dialog.appendChild(this.statusEl); // Always at the top

        // Title
        const titleEl = document.createElement('div');
        titleEl.textContent = title;
        titleEl.style.cssText = 'font-size:1.2em;font-weight:600;margin-bottom:8px;color:#e3e3e3;';
        this.dialog.appendChild(titleEl);

        // Fields
        this.fieldsEl = document.createElement('div');
        this.fieldsEl.className = 'fields';
        this._inputEls = {};
        this._fieldRows = {};

        this._renderFields();

        this.dialog.appendChild(this.fieldsEl);

        // Buttons
        const btns = document.createElement('div');
        btns.className = 'dialog-buttons';

        // Helper for status
        const setStatus = (msg, isError = false) => {
            this.statusEl.textContent = msg || '';
            this.statusEl.style.color = isError ? '#ff9800' : '#e3e3e3';
        };
        this._setStatus = setStatus;

        // Custom buttons
        for (const btn of buttons) {
            const b = document.createElement('button');
            b.textContent = btn.name;
            b.onclick = async () => {
                const values = this._collectValues();
                await btn.onClick?.(values, this, setStatus);
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

    _renderFields() {
        this.fieldsEl.innerHTML = '';
        this._inputEls = {};
        this._fieldRows = {};

        for (const field of this._fields) {
            if (field.type === 'select' && Array.isArray(field.options)) {
                if (this._values[field.name] === undefined) {
                    if (field.default !== undefined) {
                        this._values[field.name] = field.default;
                    } else if (field.options.length > 0) {
                        this._values[field.name] = field.options[0].value;
                    } else {
                        this._values[field.name] = '';
                    }
                }
            }
        }

        for (const field of this._fields) {
            // Visibility check
            if (!this._isFieldVisible(field)) continue;

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
                let initialValue = this._values[field.name];
                input.options = field.options.map(opt => ({
                    value: opt.value,
                    label: opt.label,
                    selected: initialValue === opt.value
                }));
                input.value = initialValue;
                input.addEventListener('change', () => this._onFieldChange(field.name));
            } else if (field.type === 'checkbox') {
                input = document.createElement('ui-checkbox');
                input.id = field.name;
                input.checked = !!(this._values[field.name] ?? field.default);
                input.addEventListener('change', () => this._onFieldChange(field.name));
                row.appendChild(input);
                row.appendChild(label);
                this._inputEls[field.name] = input;
                this._fieldRows[field.name] = row;
                this.fieldsEl.appendChild(row);
                continue;
            } else if (field.type === 'number' && field.integer) {
                // Custom spinner for integer number fields
                const wrapper = document.createElement('div');
                wrapper.className = 'input-spinner-wrapper';
                input = document.createElement('input');
                input.type = 'number';
                input.id = field.name;
                if (field.min !== undefined) input.min = field.min;
                if (field.step !== undefined) input.step = field.step;
                input.value = this._values[field.name] ?? field.default ?? '';
                input.addEventListener('input', () => {
                    let val = input.value;
                    val = val.replace(/[^\d]/g, '');
                    if (val === '') val = '';
                    else val = String(Math.max(field.min ?? 0, parseInt(val, 10)));
                    input.value = val;
                    this._onFieldChange(field.name);
                });

                // Spinner arrows
                const arrows = document.createElement('div');
                arrows.className = 'input-spinner-arrows';

                const up = document.createElement('div');
                up.className = 'input-spinner-arrow up img-button';
                up.tabIndex = 0;
                up.addEventListener('mousedown', e => {
                    e.preventDefault();
                    let v = parseInt(input.value || '0', 10);
                    v = isNaN(v) ? (field.min ?? 0) : v + (field.step ? parseInt(field.step, 10) : 1);
                    if (field.max !== undefined) v = Math.min(v, field.max);
                    input.value = String(v);
                    this._onFieldChange(field.name);
                });

                const down = document.createElement('div');
                down.className = 'input-spinner-arrow down img-button';
                down.tabIndex = 0;
                down.addEventListener('mousedown', e => {
                    e.preventDefault();
                    let v = parseInt(input.value || '0', 10);
                    v = isNaN(v) ? (field.min ?? 0) : v - (field.step ? parseInt(field.step, 10) : 1);
                    if (field.min !== undefined) v = Math.max(v, field.min);
                    else v = Math.max(v, 0);
                    input.value = String(v);
                    this._onFieldChange(field.name);
                });

                arrows.appendChild(up);
                arrows.appendChild(down);
                wrapper.appendChild(input);
                wrapper.appendChild(arrows);

                row.appendChild(label);
                row.appendChild(wrapper);
                this._inputEls[field.name] = input;
                this._fieldRows[field.name] = row;
                this.fieldsEl.appendChild(row);
                continue;
            } else {
                input = document.createElement('input');
                if (field.type === 'number') {
                    input.type = 'number';
                    if (field.min !== undefined) input.min = field.min;
                    if (field.step !== undefined) input.step = field.step;
                    input.value = this._values[field.name] ?? field.default ?? '';
                    input.addEventListener('input', () => this._onFieldChange(field.name));
                } else {
                    input.type = field.type === 'password' ? 'password' : 'text';
                    input.value = this._values[field.name] ?? field.default ?? '';
                    input.addEventListener('input', () => this._onFieldChange(field.name));
                }
            }
            row.appendChild(label);
            row.appendChild(input);
            this._inputEls[field.name] = input;
            this._fieldRows[field.name] = row;
            this.fieldsEl.appendChild(row);
        }
    }

    _onFieldChange(changedFieldName) {
        for (const field of this._fields) {
            const el = this._inputEls[field.name];
            if (!el) continue;
            if (field.type === 'checkbox') {
                this._values[field.name] = el.checked;
            } else {
                this._values[field.name] = el.value;
            }
        }
        // Only re-render if the changed field is a checkbox or select
        const changedField = this._fields.find(f => f.name === changedFieldName);
        if (changedField && (changedField.type === 'checkbox' || changedField.type === 'select')) {
            this._renderFields();
        }
    }

    _isFieldVisible(field) {
        if (!field.visibleIf) return true;
        // visibleIf: { fieldName: value } or function(values) => bool
        if (typeof field.visibleIf === 'function') {
            return field.visibleIf(this._values);
        }
        if (typeof field.visibleIf === 'object') {
            return Object.entries(field.visibleIf).every(([dep, val]) => this._values[dep] === val);
        }
        return true;
    }

    _collectValues() {
        const result = {};
        for (const field of this._fields) {
            if (!this._isFieldVisible(field)) continue;
            const el = this._inputEls[field.name];
            if (!el) continue;
            if (field.type === 'checkbox') {
                result[field.name] = el.checked;
            } else if (field.type === 'number') {
                result[field.name] = Number(el.value); // Convert to number
            } else {
                result[field.name] = el.value;
            }
        }
        return result;
    }

    _ok() {
        // Validate required (only visible fields)
        for (const field of this._fields) {
            if (!this._isFieldVisible(field)) continue;
            if (field.required) {
                const val = this._inputEls[field.name];
                if (field.type === 'checkbox') continue;
                if (!val || !val.value) {
                    this._setStatus(`Field "${field.label}" is required`, true);
                    return;
                }
            }
            // Validate number fields
            if (field.type === 'number') {
                const val = this._inputEls[field.name];
                if (val && val.value !== '') {
                    const num = Number(val.value);
                    if (isNaN(num) || (field.integer && !Number.isInteger(num))) {
                        this._setStatus(`Field "${field.label}" must be an integer`, true);
                        return;
                    }
                    if (field.min !== undefined && num < field.min) {
                        this._setStatus(`Field "${field.label}" must be at least ${field.min}`, true);
                        return;
                    }
                }
            }
        }
        // Custom validation
        if (typeof this._validate === 'function') {
            const err = this._validate(this._collectValues());
            if (err) {
                this._setStatus(err, true);
                return;
            }
        }
        this._resolve(this._collectValues());
        if (typeof this._onClose === 'function') this._onClose();
        this.remove();
    }

    _cancel() {
        this._resolve(null);
        if (typeof this._onClose === 'function') this._onClose();
        this.remove();
    }

    _showError(msg) {
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

window.showEditDialog = function ({ title, fields, values, buttons, validate, onClose }) {
    return new Promise(resolve => {
        const dlg = document.createElement('edit-dialog');
        document.body.appendChild(dlg);
        dlg.open({ title, fields, values, buttons, validate, onClose }).then(resolve);
    });
};