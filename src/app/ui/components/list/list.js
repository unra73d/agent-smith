'use strict'



class List extends HTMLElement {
    constructor() {
        super()
        const shadowRoot = this.attachShadow({ mode: 'open' })

        this.list = document.createElement('div')
        shadowRoot.appendChild(this.list)

        this.items = []
        monitor(this, 'items', 'list:change')
        document.addEventListener('list:change', e => this.updateList())
        this._initStyle()
    }

    async _initStyle() {
        this.shadowRoot.adoptedStyleSheets = [
            await loadCSS('global.css'),
            await loadCSS('components/list/list.css')
        ];
    }

    getItem(data) {
        const item = document.createElement('div')
        if (data.id) {
            item.setAttribute("data-id", data.id)
        } else {
            item.setAttribute("data-id", crypto.randomUUID())
        }

        let text = ''
        if (data.name) text = data.name
        if (data.text) text = data.text
        if (data.content) text = data.content
        if (data.description) text = data.description
        if (data.summary) text = data.summary

        item.innerHTML = `
            <span class="item-text">${text}</span>
        `;

        item.addEventListener('click', e => this.onItemClick(item, data))
        return item
    }

    onItemClick(item, data) { }

    updateList() {
        if (this.items) {
            this.list.innerHTML = '';

            const fragment = document.createDocumentFragment()

            for (let data of this.items) {
                const item = this.getItem(data)
                item.classList.add('list-item')
                fragment.appendChild(item)
            }

            this.list.appendChild(fragment)
        }
    }


}

customElements.define('text-list', List)

