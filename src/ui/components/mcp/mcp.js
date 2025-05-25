class MCPList extends List {
    constructor() {
        super();
        this.testMCPController = null

        document.addEventListener('storage:mcps', e => this.items = e.detail);
        document.addEventListener('mcps:new', async e => { this.handleCreateMCP() });
        this._initStyle()
    }

    async _initStyle() {
        await super._initStyle()
        this.shadowRoot.adoptedStyleSheets = [
            ...this.shadowRoot.adoptedStyleSheets,
            await loadCSS('components/mcp/mcp.css')
        ];
    }

    async showMCPDialog({ title, initialValues = {}, validate, onSave }) {
        const fields = [
            { name: 'name', label: 'MCP name', type: 'text', required: true },
            {
                name: 'transport',
                label: 'Type',
                type: 'select',
                required: true,
                options: [
                    { value: 'sse', label: 'SSE' },
                    { value: 'stdio', label: 'stdio' }
                ]
            },
            {
                name: 'url',
                label: 'URL endpoint',
                type: 'text',
                required: true,
                visibleIf: { transport: 'sse' }
            },
            {
                name: 'command',
                label: 'Command',
                type: 'text',
                required: true,
                visibleIf: { transport: 'stdio' }
            }
        ];

        const buttons = [
            {
                name: 'Test MCP',
                onClick: async (values, dialog, setStatus) => {
                    if (this.testMCPController) {
                        this.testMCPController.abort()
                        this.testMCPController = null
                    }
                    this.testMCPController = new AbortController();

                    setStatus('Testing MCP...', false);
                    try {
                        const ok = await apiMCPTest(values, this.testMCPController.signal);
                        if (ok != 'canceled') {
                            if (ok) {
                                setStatus('MCP test successful!', false);
                            } else {
                                setStatus('MCP test failed.', true);
                            }
                        }
                    } catch (err) {
                        setStatus('Error testing MCP: ' + (err.message || err), true);
                    }
                }
            }
        ];

        let res = await showEditDialog({
            title,
            fields,
            validate,
            buttons,
            values: initialValues,
            onClose: () => {
                if (this.testMCPController) {
                    this.testMCPController.abort()
                    this.testMCPController = null
                }
            }
        });

        if (res) {
            await onSave(res);
        }
    }

    async handleCreateMCP() {
        const existingNames = (this.items || []).map(item => item.name);
        const validate = values => {
            if (existingNames.includes(values.name)) {
                return 'MCP name must be unique';
            }
        };
        await this.showMCPDialog({
            title: 'New MCP server',
            initialValues: {},
            validate,
            onSave: async (res) => {
                res.active = true
                await apiMCPCreate(res);
            }
        });
    }

    async handleEditMCP(mcp) {
        const existingNames = (this.items || []).filter(item => item.id !== mcp.id).map(item => item.name);
        const validate = values => {
            if (existingNames.includes(values.name)) {
                return 'MCP name must be unique';
            }
        };
        await this.showMCPDialog({
            title: 'Edit MCP server',
            initialValues: mcp,
            validate,
            onSave: async (res) => {
                res.id = mcp.id;
                res.active = mcp.active
                await apiMCPUpdate(res);
            }
        });
    }

    getItem(data) {
        const item = document.createElement('div')

        item.innerHTML = `
            <div class="item-header">
                <ui-checkbox class="select-all-checkbox" label="${data.loaded ? '' : '(Loading)'}${data.name}" ${data.active ? 'checked' : ''}></ui-checkbox>
                <div alt="Edit" class="edit-icon img-button" data-id="${data.id}">*</div>
                <div alt="Delete" class="delete-icon img-button" data-id="${data.id}">&#xe053;</div>
            </div>
            <div class="item-content"></div>
        `;

        const selectAllCheckbox = item.querySelector('ui-checkbox');
        const editIcon = item.querySelector('.edit-icon');
        const deleteIcon = item.querySelector('.delete-icon');
        const itemContent = item.querySelector('.item-content');

        if (!data.active) itemContent.classList.add('disabled')
        for (let tool of data.tools) {
            const toolItem = document.createElement('div');
            toolItem.classList.add('tool-item');
            toolItem.innerHTML = `
                <span class="item-text">${tool.name} - ${tool.description}</span>
            `;
            itemContent.appendChild(toolItem);

        }

        selectAllCheckbox.addEventListener('change', async e => {
            data.active = e.target.checked
            await apiMCPUpdate(data)
        });

        editIcon.addEventListener('click', e => this.handleEditMCP(data));
        deleteIcon.addEventListener('click', e => this.handleDeleteItem(data.id));

        return item;
    }

    async handleDeleteItem(itemId) {
        const confirmed = await confirmDialog('Are you sure you want to delete this item? This action cannot be undone.');

        if (confirmed) {
            apiMCPDelete(itemId)
        }
    }
}

customElements.define('mcp-list', MCPList);