'use strict'

class RoleList extends List {
    constructor() {
        super();

        document.addEventListener('storage:roles', e => this.items = Storage.roles || []);
        document.addEventListener('roles:new', async e => { this.createNewRole() });

        this._initStyle();
    }

    async _initStyle() {
        await super._initStyle();
        this.shadowRoot.adoptedStyleSheets = [
            ...this.shadowRoot.adoptedStyleSheets,
            await loadCSS('components/roles/roles.css')
        ];
    }

    getItem(role) {
        const item = document.createElement('div');

        item.innerHTML = `
            <div class="role-header">
                <span class="role-name">${role.config.name}</span>
                <div alt="Edit" class="edit-icon img-button" data-id="${role.id}">*</div>
                <div alt="Delete" class="delete-icon img-button" data-id="${role.id}">&#xe053;</div>
            </div>
            <div class="role-content">
                <p>General instruction:</p>
                <div class="role-item">
                    ${role.config.generalInstruction}
                </div>
                <p>Role & personality:</p>
                <div class="role-item">
                    ${role.config.role}
                </div>
                <p>Conversation style & tone:</p>
                <div class="role-item">
                    ${role.config.style}
                </div>
            </div>
        `;

        item.querySelector('.delete-icon').addEventListener('click', e => this.handleDelete(e, role.id));
        item.querySelector('.edit-icon').addEventListener('click', e => this.handleEdit(e, role));

        return item;
    }

    async handleDelete(e, roleId) {
        const confirmed = await confirmDialog('Delete this agent role?');
        if (confirmed) {
            await apiDeleteRole(roleId);
        }
    }

    async showRoleDialog({ title, initialValues, onSave }) {
        const fields = [
            { name: 'name', label: 'Name', type: 'text', required: true },
            { name: 'generalInstruction', label: 'General instruction', type: 'text', required: false, multiline: true },
            { name: 'role', label: 'AI role and personality', type: 'text', required: false, multiline: true },
            { name: 'style', label: 'Conversation style & tone', type: 'text', required: false, multiline: true },
        ];
        const res = await showEditDialog({
            title,
            fields,
            values: initialValues,
            buttons: [],
            onClose: () => { }
        });

        if (res) {
            await onSave(res);
        }
    }

    async handleEdit(e, role) {
        await this.showRoleDialog({
            title: 'Edit Role',
            initialValues: role.config,
            onSave: async (res) => {
                res["id"] = role.id;
                await apiUpdateRole(res);
            }
        });
    }

    async createNewRole() {
        await this.showRoleDialog({
            title: 'New Role',
            initialValues: {},
            onSave: async (res) => {
                await apiCreateRole(res);
            }
        });
    }
}

customElements.define('role-list', RoleList);