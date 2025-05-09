'use strict'

class List extends HTMLElement {
    constructor() {
        super()
        const shadowRoot = this.attachShadow({mode: 'open'})
        
        const styles = document.createElement('style');
        styles.innerHTML = `
        @import url('global.css');
        @import url('components/list/list.css');
        `
        shadowRoot.appendChild(styles);

        this.list = document.createElement('div')
        shadowRoot.appendChild(this.list)

        this.items = []
        monitor(this, 'items', 'list:change')
        document.addEventListener('list:change', e=>this.updateList())
    }

    getItem(data){
        const item = document.createElement('div')
        if(data.id){
            item.setAttribute("data-id", data.id)
        } else {
            item.setAttribute("data-id", crypto.randomUUID())
        }

        let text = ''
        if(data.name)text = data.name
        if(data.text)text = data.text
        if(data.content)text = data.content
        if(data.description)text = data.description
        if(data.summary)text = data.summary

        item.innerHTML = `
            <span class="item-text">${text}</span>
        `;

        item.addEventListener('click', e=>this.onItemClick(item, data))
        return item
    }

    onItemClick(item, data){}

    updateList(){
        if(this.items && this.items.length > 0){
            this.list.innerHTML = ''

            for (let data of this.items) {
                this._appendItem(data)
            }
        }
    }

    _appendItem(data, front=false){
        const item = this.getItem(data)
        item.classList.add('list-item')
        
        if(front){
            this.list.insertBefore(item, this.list.firstChild);
        } else {
            this.list.appendChild(item);
        }
    }


}

customElements.define('text-list', List)

