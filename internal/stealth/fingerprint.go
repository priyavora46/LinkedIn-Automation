package stealth

import (
	"fmt"
	"math/rand"

	"github.com/go-rod/rod"
	"github.com/go-rod/rod/lib/proto"
)

var userAgents = []string{
	"Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/120.0.0.0 Safari/537.36",
	"Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/119.0.0.0 Safari/537.36",
	"Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/120.0.0.0 Safari/537.36",
	"Mozilla/5.0 (X11; Linux x86_64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/120.0.0.0 Safari/537.36",
}

// RandomUserAgent returns a random user agent string
func RandomUserAgent() string {
	return userAgents[rand.Intn(len(userAgents))]
}

// MaskWebDriver removes automation detection flags
func MaskWebDriver(page *rod.Page) error {
	// Override navigator.webdriver
	script := `
		() => {
			// Webdriver flag
			Object.defineProperty(navigator, 'webdriver', {
				get: () => false
			});
			
			// Plugins
			Object.defineProperty(navigator, 'plugins', {
				get: () => [
					{
						0: {type: "application/x-google-chrome-pdf", suffixes: "pdf", description: "Portable Document Format"},
						description: "Portable Document Format",
						filename: "internal-pdf-viewer",
						length: 1,
						name: "Chrome PDF Plugin"
					},
					{
						0: {type: "application/pdf", suffixes: "pdf", description: "Portable Document Format"},
						description: "Portable Document Format",
						filename: "mhjfbmdgcfjbbpaeojofohoefgiehjai",
						length: 1,
						name: "Chrome PDF Viewer"
					}
				]
			});
			
			// Languages
			Object.defineProperty(navigator, 'languages', {
				get: () => ['en-US', 'en']
			});
			
			// Platform
			Object.defineProperty(navigator, 'platform', {
				get: () => 'Win32'
			});
			
			// Chrome runtime
			window.chrome = {
				runtime: {},
				loadTimes: function() {},
				csi: function() {},
				app: {}
			};
			
			// Permissions
			const originalQuery = window.navigator.permissions.query;
			window.navigator.permissions.query = (parameters) => (
				parameters.name === 'notifications' ?
					Promise.resolve({ state: Notification.permission }) :
					originalQuery(parameters)
			);
			
			// Media devices
			Object.defineProperty(navigator, 'mediaDevices', {
				get: () => ({
					enumerateDevices: () => Promise.resolve([
						{deviceId: "default", kind: "audioinput", label: "Default", groupId: "default"},
						{deviceId: "default", kind: "audiooutput", label: "Default", groupId: "default"},
						{deviceId: "default", kind: "videoinput", label: "Default", groupId: "default"}
					])
				})
			});
			
			// Battery
			Object.defineProperty(navigator, 'getBattery', {
				get: () => () => Promise.resolve({
					charging: true,
					chargingTime: 0,
					dischargingTime: Infinity,
					level: 1
				})
			});
			
			// Connection
			Object.defineProperty(navigator, 'connection', {
				get: () => ({
					effectiveType: '4g',
					rtt: 50,
					downlink: 10,
					saveData: false
				})
			});
		}
	`

	_, err := page.Eval(script)
	return err
}

// RandomizeViewport sets a random but realistic viewport size
func RandomizeViewport(page *rod.Page) error {
	viewports := []struct{ width, height int }{
		{1920, 1080},
		{1366, 768},
		{1536, 864},
		{1440, 900},
		{1280, 720},
	}

	vp := viewports[rand.Intn(len(viewports))]
	return page.SetViewport(&proto.EmulationSetDeviceMetricsOverride{
		Width:             vp.width,
		Height:            vp.height,
		DeviceScaleFactor: 1,
		Mobile:            false,
	})
}

// AddRandomCanvas adds canvas noise to prevent fingerprinting
func AddRandomCanvas(page *rod.Page) error {
	script := fmt.Sprintf(`
		() => {
			const originalToDataURL = HTMLCanvasElement.prototype.toDataURL;
			const originalToBlob = HTMLCanvasElement.prototype.toBlob;
			const noise = () => Math.random() * %f;
			
			HTMLCanvasElement.prototype.toDataURL = function() {
				const context = this.getContext('2d');
				const imageData = context.getImageData(0, 0, this.width, this.height);
				for (let i = 0; i < imageData.data.length; i += 4) {
					imageData.data[i] = Math.min(255, imageData.data[i] + noise());
				}
				context.putImageData(imageData, 0, 0);
				return originalToDataURL.apply(this, arguments);
			};
		}
	`, 0.1+rand.Float64()*0.5)

	_, err := page.Eval(script)
	return err
}

// SpoofTimezone sets a realistic timezone
