const confirmOverlay = document.getElementById('confirmOverlay');
const confirmMessageEl = document.getElementById('confirmMessage');
const confirmYesBtn = document.getElementById('confirmYes');
const confirmNoBtn = document.getElementById('confirmNo');

function confirmDialog(message) {
    return new Promise((resolve) => {
        confirmMessageEl.textContent = message;
        confirmOverlay.style.display = 'flex';

        const yesListener = () => {
            cleanup();
            resolve(true);
        };

        const noListener = () => {
            cleanup();
            resolve(false);
        };

        // Function to remove listeners and hide dialog
        const cleanup = () => {
            confirmYesBtn.removeEventListener('click', yesListener);
            confirmNoBtn.removeEventListener('click', noListener);
            confirmOverlay.removeEventListener('click', backgroundClickListener);
            confirmOverlay.style.display = 'none';
        };

        // Close dialog if clicking outside the dialog box
        const backgroundClickListener = (event) => {
            if (event.target === confirmOverlay) {
                noListener();
            }
        };

        // Add listeners
        confirmYesBtn.addEventListener('click', yesListener);
        confirmNoBtn.addEventListener('click', noListener);
        confirmOverlay.addEventListener('click', backgroundClickListener);

    });
}
function initShortcuts() {
    document.addEventListener('keydown', (event) => {
        const isModifier = event.metaKey || event.ctrlKey;

        if (!isModifier) {
            return;
        }

        let command = null;
        switch (event.key.toLowerCase()) {
            case 'c':
                command = 'copy';
                break;
            case 'v':
                command = 'paste';
                break;
            case 'x':
                command = 'cut';
                break;
            case 'a':
                command = 'selectAll';
                break;
            case 'z':
                if (event.shiftKey) {
                    command = 'redo';
                } else {
                    command = 'undo';
                }
                break
            case 'y':
                command = 'redo';
                break;
        }

        if (command) {
            try {
                if (document.execCommand(command)) {
                    event.preventDefault();
                } else {
                    console.warn(`document.execCommand('${command}') failed or did not apply.`);
                }
            } catch (e) {
                console.error(`Error executing document.execCommand('${command}'):`, e);
            }
        }
    });
}

function splitStringBySearchText(inputString, searchText) {
    const index = inputString.indexOf(searchText);
    if (index === -1) {
        return [inputString, ''];
    }
    return [inputString.slice(0, index), inputString.slice(index + searchText.length)];
}

function sendEvent(name, data = {}) {
    e = new CustomEvent(name, { detail: data });
    document.dispatchEvent(e)
    console.debug(name)
}

function monitor(obj, prop, eventName) {
    let value = obj[prop]
    Object.defineProperty(obj, prop, {
        get: function () {
            return value;
        },
        set: function (newValue) {
            value = newValue;
            sendEvent(eventName, value)
        },
        enumerable: true,
        configurable: true
    });
}


function deepMonitor(obj, propName, eventName) {
    // WeakMap to cache proxies: maps an original object to its proxy created by this monitor instance.
    // This serves two purposes:
    // 1. Prevents re-proxying the same object multiple times if it appears in different parts of the
    //    monitored structure or is re-accessed.
    // 2. Allows retrieval of the existing proxy if the original object is encountered again.
    const proxyCache = new WeakMap();

    /**
     * Recursively creates a Proxy for the target object and its nested objects/arrays.
     * Operations on these proxies will trigger the rootEventCallback.
     * @param {*} target The value to potentially proxy.
     * @param {function} rootEventCallback The function to call when a change is detected anywhere
     * within the proxied structure.
     * @returns {*} The proxied version of the target if it's an object/array, otherwise the target itself.
     */
    const createDeepProxy = (target, rootEventCallback) => {
        // If target is not an object or array, or is null, it cannot/should not be proxied.
        if (typeof target !== 'object' || target === null) {
            return target;
        }

        // If this original object already has a proxy created by this specific monitor instance,
        // return that existing proxy to ensure consistency and avoid redundant proxying.
        if (proxyCache.has(target)) {
            return proxyCache.get(target);
        }

        const handler = {
            get(currentTarget, key, receiver) {
                // Reflect.get retrieves the property value from the target.
                const result = Reflect.get(currentTarget, key, receiver);

                // Recursively create deep proxies for nested properties upon access.
                // If the accessed property is an object or array, it will also be wrapped in a proxy.
                // This ensures that changes to obj.prop.nested.value are also monitored.
                // Functions (which are typeof 'object') are generally not proxied further by this logic,
                // as we are interested in data changes. Methods are returned as-is; their execution
                // on the proxy will trigger 'set' or 'deleteProperty' traps if they modify the object.
                if (typeof result === 'object' && result !== null) {
                    return createDeepProxy(result, rootEventCallback);
                }
                return result;
            },
            set(currentTarget, key, value, receiver) {
                const oldValue = currentTarget[key];
                // Reflect.set applies the change to the target.
                const success = Reflect.set(currentTarget, key, value, receiver);

                // Determine if the value actually changed to avoid redundant event triggers.
                let changed = true;
                if (oldValue === value) { // Handles primitives and cases where the same object reference is set
                    changed = false;
                } else if (Number.isNaN(oldValue) && Number.isNaN(value)) { // Both are NaN, consider them unchanged relative to each other
                    changed = false;
                }
                // This `set` trap is triggered by:
                // 1. Direct property assignments on an object (e.g., obj.prop.name = "new").
                // 2. Array element assignments by index (e.g., obj.prop[0] = "new").
                // 3. Array length modifications (e.g., obj.prop.length = x, or implicitly by methods like push/pop).
                //    When array methods (e.g., proxy.push(item)) are called:
                //    a) 'get' trap provides the original method.
                //    b) The method executes with 'this' as the proxy.
                //    c) Internal operations of the method (like setting new indices or updating 'length')
                //       are intercepted by this 'set' trap.

                if (changed && success) { // Trigger event only if value changed and set was successful
                    rootEventCallback();
                }
                return success;
            },
            deleteProperty(currentTarget, key, receiver) {
                // Reflect.deleteProperty applies the deletion to the target.
                const success = Reflect.deleteProperty(currentTarget, key, receiver);
                // Always trigger event on successful deletion, as it's a definite change.
                if (success) {
                    rootEventCallback();
                }
                return success;
            }
        };

        // Create the proxy for the current target object/array.
        const proxy = new Proxy(target, handler);
        // Cache the proxy, mapping the original target to this new proxy.
        proxyCache.set(target, proxy);
        return proxy;
    };

    // This callback is invoked whenever a change is detected:
    // - within a deeply nested property of the monitored object/array, OR
    // - when the top-level monitored property (obj[propName]) itself is reassigned.
    const triggerRootEventCallback = () => {
        // obj[propName] will use the Object.defineProperty getter defined below.
        // This getter returns currentPropertyValueHolder, which is the
        // current, (potentially) proxied, value of the monitored property.
        // This ensures sendEvent receives the correct top-level monitored property value.
        sendEvent(eventName, obj[propName]);
    };

    // This variable holds the actual current value of obj[propName].
    // It might be a primitive, or it will be a Proxy if the original value was an object/array.
    let currentPropertyValueHolder = obj[propName];

    // Initial setup: If the property's current value is an object or array,
    // replace currentPropertyValueHolder with its deeply proxied version.
    // This ensures that changes to an initially empty object/array that is later populated are caught.
    currentPropertyValueHolder = createDeepProxy(currentPropertyValueHolder, triggerRootEventCallback);

    // Use Object.defineProperty to intercept direct get/set operations on obj[propName].
    // This primarily handles the case where the entire obj[propName] is reassigned
    // (e.g., obj.prop = newValue;).
    Object.defineProperty(obj, propName, {
        get() {
            return currentPropertyValueHolder;
        },
        set(newValue) {
            // This setter is called when an assignment like obj.propName = someNewValue; occurs.

            // The new value (newValue) also needs to be made deeply reactive if it's an object/array.
            // currentPropertyValueHolder will now hold the new (and potentially proxied) value.
            currentPropertyValueHolder = createDeepProxy(newValue, triggerRootEventCallback);

            // A reassignment of the top-level monitored property itself has occurred, so trigger the event.
            triggerRootEventCallback();
        },
        configurable: true, // Allows the property to be re-defined or deleted later if necessary.
        enumerable: true    // Ensures the property shows up in for...in loops and Object.keys().
    });
}

async function loadCSS(url) {
    if (!loadCSS.cache) {
        loadCSS.cache = {};
    }

    if (loadCSS.cache[url]) {
        return loadCSS.cache[url];
    }

    try {
        const response = await fetch(url);
        if (!response.ok) {
            throw new Error(`Failed to load CSS from ${url}: ${response.status}`);
        }
        const text = await response.text();
        const sheet = new CSSStyleSheet();
        sheet.replace(text);
        loadCSS.cache[url] = sheet;
        return sheet;
    } catch (error) {
        console.error(error);
        return null;
    }
}

// const gStyleSheet = await loadCSS('global.css')

initShortcuts();