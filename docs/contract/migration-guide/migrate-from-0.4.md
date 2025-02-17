---
order: 2
---

# Migrating from Solidity++ 0.4

Solidity++ 0.8 is syntactically closer to Solidity than Solidity++ 0.4, which is more friendly to developers from Solidity; some counterintuitive syntax and keywords are removed in this version.


## Solidity++ 0.8.0 Changes
### Breaking Changes
* All Solidity breaking changes since v0.4.3 to v0.8.0 are applied to Solidity++ 0.8.0 such as: ABI coder v2 [on by default](https://docs.soliditylang.org/en/latest/080-breaking-changes.html), `fallback` and `receive` [keywords, and conversions](https://docs.soliditylang.org/en/latest/060-breaking-changes.html#semantic-and-syntactic-changes) between `address` and `address payable`.
* Most Solidity contracts and libraries can be compiled by solppc 0.8.0 and imported by Solidity++ 0.8.0 contracts, except for the contracts with external / delegate calls.
* `onMessage` keyword is deprecated since 0.8.0. Use function declarations instead.
* `message` keyword is deprecated since 0.8.0. Use `function` declared in a `contract` or `interface` instead.
* `send` keyword is deprecated since 0.8.0. Use function calls instead.
* `getter` keyword is deprecated since 0.8.0. Use `view`/`pure` functions instead.
* Inline assembly and Yul are available since 0.8.0.
* `keccak256` is available since 0.8.0.
* `tokenId` keyword (Vite Native Token Id type) is changed to `vitetoken`.
* Some transaction properties are changed since 0.8.0: `msg.tokenid` is changed to `msg.token`, `msg.amount` is changed to `msg.value`.

### Compatibilities
* An external or public `function` (without return values) in 0.8.0 equivalents to `onMessage` in 0.4.3. They have the same signature and selector. For example `function set(uint data) external` equivalents to `onMessage set(uint data)`.
* An external or public `function` declaration in 0.8.0 equivalents to `message` declaration in 0.4.3. 
* A contract compiled by solppc 0.8.0 can call contracts compiled by solppc 0.4.3 asynchronously.
* A contract compiled by solppc 0.4.3 can call contracts compiled by solppc 0.8.0 asynchronously.
* The ABI encoding, storage layout, memory layout and calldata layout remain unchanged.

Also available starting in 0.8.0 is passing function as callbacks in parameters.
```javascript
// SPDX-License-Identifier: GPL-3.0
pragma soliditypp ^0.8.0;

contract A {
    function add(uint a, uint b, function(uint) external callback) external {
        if (callback.address != address(0)) {
            // send callback to return data to the caller
            callback(a + b);
        }
    }
}

contract B {
    A contractA;
    uint public data;

    constructor (address addr) {
        contractA = A(addr);
    }

    function test() external {
        contractA.add(1, 2, this.callback_onAdd);
    }

    function callback_onAdd(uint result) external {
        // receive data from the called contract
        require(msg.sender == address(contractA));
        data = result;
    }
}
```

## Migrating to 0.8.0

Let's start with an example:

```javascript
pragma soliditypp ^0.4.3;

contract A {
    message sum(uint result);
    
    onMessage add(uint a, uint b) {
        uint result = a + b;
        address sender = msg.sender;
        send(sender, sum(result));
   }
}

contract B {
    address addrA;
    uint total;
    message add(uint a, uint b);

    constructor (address addr) {
        addrA = addr;
    }

    onMessage invoke(uint a, uint b) {
        send(addrA, add(a, b));
    }

    onMessage sum(uint result) {
        total += result;
    }

    getter total() returns(uint) {
        return total;
    }

    getter getSomething() returns(uint) {
        return total + 1;
    }
}
```

In above code, contract A declares a message listener `add(uint a, uint b)`.

contract B declares `add` message which has the same signature to `add` message listener in contract A.

Contract B declares a message listener `invoke` as the entry to the contract. When `B.invoke()` is called, contract B sends a `add` message to contract A to initiate an asynchronous message call.

When contract A responds to the message call, it sends a `sum` message back to contract B to return data asynchronously.

Contract B also declares a message listener `sum(uint result)` as a *'callback function'* to handle the returned message from contract A.

Since 0.8.0, `onMessage` and `message` are replaced with `function` and `send` statements are replaced with function calls. The `await` operator is not allowed in 0.8.0, it is required to declare callbacks explicitly using `function`.

The migrated code in 0.8.0 is as follows:

```javascript
// SPDX-License-Identifier: GPL-3.0
pragma soliditypp ^0.8.0;

interface Listener {
    // delare a callback to receive the result
    function sum(uint result) external;
}

contract A {
    // the onMessage is replaced with an external function
    function add(uint a, uint b) external {
        Listener sender = Listener(msg.sender);
        // send message to the caller
        sender.sum(a + b);
    }
}

contract B is Listener {
    A contractA;
    uint public total;  // a getter function will be auto-generated by the compiler

    constructor (address addr) {
        contractA = A(addr);
    }

    // the onMessage is replaced with a function
    function invoke(uint a, uint b) external {
        // replace the send statement with a function call
        contractA.add(a, b);
    }
    // the callback is replaced with a function
    function sum(uint result) external override {
        total += result;
    }
    // the offchain getter is replaced with a view or pure function
    function getSomething() external view returns(uint) {
        return total + 1;
    }
}
```

## Solidity++ 0.8.1 Changes

**Note**: 0.8.1 contracts do not yet compile with the [Solidity++ 0.8 Preview VS Code extension](https://marketplace.visualstudio.com/items?itemName=ViteLabs.solppdebugger)

### Breaking Changes
* Solidity contracts and libraries with external / delegate calls can be compiled by solppc 0.8.1 and imported by Solidity++ 0.8.1 contracts.
* `await` operator is introduced since 0.8.1.
* Error propagation through `revert` and `try/catch` is enabled since 0.8.1.
* Delegate calls are allowed since 0.8.1.
* External / delegate calls in Solidity are allowed since 0.8.1.
* Library linking and calls to external library functions are allowed since 0.8.1.

### Compatibilities
* A contract compiled by solppc 0.8.1 can call contracts compiled by solppc 0.4.3 asynchronously.
* A contract compiled by solppc 0.8.1 can call contracts compiled by solppc 0.8.0 both synchronously and asynchronously.
* A contract compiled by solppc 0.8.1 can get return values after a synchronous call to a contract compiled by solppc 0.8.0.
* The ABI encoding, storage layout, memory layout and calldata layout remain unchanged.

## Migrating to 0.8.1

Since Solidity++ 0.8.1, no explicit callback declarations are required.

The compiler is smart enough to generate callbacks automatically. The code is simplified and optimized significantly by `await` syntactic sugar.

The migrated code in 0.8.1 is as follows:

```javascript
// SPDX-License-Identifier: GPL-3.0
pragma soliditypp ^0.8.1;

contract A {
    // the async function can return data to the caller
    function add(uint a, uint b) external async returns(uint) {
        return a + b;
    }
}

contract B {
    A contractA;
    uint public total;

    constructor (address addr) {
        contractA = A(addr);
    }

    function invoke(uint a, uint b) external async {
        // use await expression to get data returned from the called contract
        total += await contractA.add(a, b);
    }

    function getSomething() external view returns(uint) {
        return total + 1;
    }
}
```



