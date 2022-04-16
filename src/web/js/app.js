const buttons = document.createElement('div');
document.body.appendChild(buttons);

const messages = document.createElement('ul');
document.body.appendChild(messages);

let ws;

const addButton = (text, command) => {
  const button = document.createElement('button');
  buttons.appendChild(button);
  button.innerText = text;
  button.ontouchstart = button.onmousedown = (event) => {
    ws.send(command);
  };
  button.ontouchend = button.onmouseup = (event) => {
    ws.send('stop');
  };
};

const addMessage = (text) => {
  const message = document.createElement('li');
  const content = document.createTextNode(text);
  message.appendChild(content);
  messages.prepend(message);
};

const attemptConnect = () => {
  ws = new WebSocket("ws://" + location.host + "/ws");

  ws.onclose = (event) => {
    addMessage('close');
    ws.close();
    addMessage('try again in 2s...');
    setTimeout(() => {
      attemptConnect();
    }, 2000);
  }

  ws.onerror = (event) => {
    console.error('error', {
      event
    });

    addMessage('error');
  };

  ws.onopen = (event) => {
    addMessage('open');
  };

  ws.onmessage = (event) => {
    addMessage('message: ' + event.data);
  };
};

attemptConnect();

addButton('Forwards', 'forwards');
addButton('Backwards', 'backwards');
addButton('Spin Clockwise', 'spin_cw');
addButton('Spin Counter-clockwise', 'spin_ccw');
