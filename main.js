const electron = require('electron')
// Module to control application life.
const app = electron.app
// Module to create native browser window.
const bw = electron.BrowserWindow


//var cr = require('crash-reporter').start();
let mw

var fsts = require('child_process').spawn(__dirname+'/fsts');

const globalShortcut = electron.globalShortcut;


app.on('window-all-closed', function() {
  app.quit();
});


app.on('ready', function() {
    mw = new bw({width: 800,
                height: 359,
                useContentSize: true,
                skipTaskbar: true,
                resizable: true,
                icon: __dirname+'/fsto.png',
                darkTheme: true,
                frame: false,
                autosize: true,
                x: 0,
                y: 0});
    mw.setMenuBarVisibility(false);
    mw.setTitle("Fs.life!");
    mw.setAlwaysOnTop(false);
 // mw.openDevTools();
    mw.loadURL('http://localhost:8000/x');
  var wc = mw.webContents;
   //console.log(wc);
  // Emitted when the window is closed.
  var ret = globalShortcut.register('ctrl+w', function() {
    fsts.kill('SIGINT');
    app.quit();
  });
  if (!ret) {
    console.log('registration failed');
  }

  var h = globalShortcut.register('ctrl+`', function() {
  // wc.reloadIgnoringCache()
      if (mw.isVisible()){mw.hide();} else {mw.show();wc.reload();}
  });

  // Check whether a shortcut is registered.
  console.log(globalShortcut.isRegistered('ctrl+`'));
});

app.on('will-quit', function() {
  // Unregister a shortcut.
  globalShortcut.unregister('ctrl+w');
  globalShortcut.unregister('ctrl+`');

  // Unregister all shortcuts.
  globalShortcut.unregisterAll();
  mw.on('closed', function() {
    mw = null;
    if (fsts){fsts.kill('SIGINT')}
    app.quit();
  });
});
