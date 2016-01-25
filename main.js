var app = require('app');
var BrowserWindow = require('browser-window');
//var cr = require('crash-reporter').start();
var mainWindow = null;
var fsts = require('child_process').spawn(__dirname+'/fsts');

const electron = require('electron');
const globalShortcut = electron.globalShortcut;


app.on('window-all-closed', function() {
  app.quit();
});


app.on('ready', function() {
    mainWindow = new BrowserWindow({width: 750,
                                    height: 359,
                                   'use-content-size': true,
                                   'skip-taskbar': true,
                                    darkTheme:true,
                                    frame: false,
                                    autosize: 1,
                                    x: 0,
                                    y: 0});
    mainWindow.setMenuBarVisibility(false);
    mainWindow.setTitle("Fsto!");
    mainWindow.setAlwaysOnTop(false);
 // mainWindow.openDevTools();
    mainWindow.loadURL('http://localhost:8000/x');
  var wc = mainWindow.webContents;
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
      if (mainWindow.isVisible()){mainWindow.hide();} else {mainWindow.show();wc.reload();}
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
  mainWindow.on('closed', function() {
    mainWindow = null;
    if (fsts){fsts.kill('SIGINT')}
    app.quit();
  });
});
