'use strict'

class ProviderList extends List {
    constructor() {
        super();
        this.testProviderController = null

        document.addEventListener('storage:providers', e => this.items = Storage.providers || []);
        document.addEventListener('providers:new', async e => { this.createNewProvider() });

        this._initStyle();
    }

    async _initStyle() {
        await super._initStyle();
        this.shadowRoot.adoptedStyleSheets = [
            ...this.shadowRoot.adoptedStyleSheets,
            await loadCSS('components/providers/providers.css')
        ];
    }

    getItem(provider) {
        const item = document.createElement('div');

        item.innerHTML = `
            <div class="provider-header">
                <span class="provider-name">${provider.name}</span>
                <div alt="Edit" class="edit-icon img-button" data-id="${provider.id}">*</div>
                <div alt="Delete" class="delete-icon img-button" data-id="${provider.id}">&#xe053;</div>
            </div>
            <div class="provider-content">
                <div class="rate-limit">
                    <span>Rate Limit: ${provider.rateLimit}</span>
                </div>
                ${provider.models.map(model => `
                    <div class="model-item">
                        <span class="model-name">• ${model.name}</span>
                    </div>
                `).join('')}
            </div>
        `;

        item.querySelector('.delete-icon').addEventListener('click', e => this.handleDelete(e, provider.id));
        item.querySelector('.edit-icon').addEventListener('click', e => this.handleEdit(e, provider));

        return item;
    }

    async handleDelete(e, providerId) {
        const confirmed = await confirmDialog('Delete this AI provider?');
        if (confirmed) {
            await apiDeleteProvider(providerId);
        }
    }

    async showProviderDialog({ title, initialValues, onSave }) {
        const fields = [
            { name: 'name', label: 'Provider Name', type: 'text', required: true },
            { name: 'url', label: 'API URL', type: 'text', required: true },
            { name: 'apiKey', label: 'API Key', type: 'text', required: false },
            {
                name: 'rateLimit',
                label: 'Rate limit (requests per minute)',
                type: 'number',
                required: false,
                min: 0,
                step: 1,
                integer: true
            }
        ];

        const buttons = [
            {
                name: 'Test Provider',
                onClick: async (values, dialog, setStatus) => {
                    if (this.testProviderController) {
                        this.testProviderController.abort();
                        this.testProviderController = null;
                    }
                    this.testProviderController = new AbortController();

                    setStatus('Testing provider...', false);
                    try {
                        const ok = await apiTestProvider(values, this.testProviderController.signal);
                        if (ok) {
                            setStatus('Provider test successful!', false);
                        } else {
                            setStatus('Provider test failed.', true);
                        }
                    } catch (err) {
                        setStatus('Error testing provider: ' + (err.message || err), true);
                    }
                }
            }
        ];

        const res = await showEditDialog({
            title,
            fields,
            values: initialValues,
            buttons,
            onClose: () => {
                if (this.testProviderController) {
                    this.testProviderController.abort();
                    this.testProviderController = null;
                }
            }
        });

        if (res) {
            await onSave(res);
        }
    }

    async handleEdit(e, provider) {
        await this.showProviderDialog({
            title: 'Edit Provider',
            initialValues: provider,
            onSave: async (res) => {
                res["id"] = provider.id;
                await apiUpdateProvider(res);
            }
        });
    }

    async createNewProvider() {
        await this.showProviderDialog({
            title: 'New Provider',
            initialValues: {},
            onSave: async (res) => {
                await apiCreateProvider(res);
            }
        });
    }
}

customElements.define('provider-list', ProviderList);