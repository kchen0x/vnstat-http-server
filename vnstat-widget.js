// vnstat-http-server Widget for Scriptable
// Designed with reference to Cloudflare Widget style

// ==================== Configuration ====================
const CONFIG = {
  SERVER_URL: 'http://your-server-ip:8080',
  TOKEN: 'your-secret-token',
  REFRESH_INTERVAL: 300,
  INTERFACE_NAME: '',
  WIDGET_TITLE: 'Traffic Monitor',  // Custom widget title
  MONTHLY_LIMIT_GB: 1000,    // Monthly traffic limit in GB
  // Progress bar configuration
  BOX_COUNT: 10,  // Number of boxes in progress bar
  IS_SQUARE: true  // true = square, false = rectangle
};
// ========================================================

// Scale image to specified height
function scaleImage(imageSize, height) {
  const scale = height / imageSize.height;
  return new Size(scale * imageSize.width, height);
}

// Format bytes with appropriate unit
function formatBytes(bytes) {
  if (bytes === 0) return '0 B';
  const k = 1024;
  const sizes = ['B', 'KB', 'MB', 'GB', 'TB'];
  const i = Math.floor(Math.log(bytes) / Math.log(k));
  const value = bytes / Math.pow(k, i);
  if (value >= 100) {
    return value.toFixed(0) + ' ' + sizes[i];
  } else if (value >= 10) {
    return value.toFixed(1) + ' ' + sizes[i];
  } else {
    return value.toFixed(2) + ' ' + sizes[i];
  }
}

// Format number with simplified display
function formatNumber(num) {
  if (num < 1000) {
    return num.toString();
  }
  const suffixes = ['', 'k', 'm', 'b', 't'];
  let suffixIndex = 0;
  let formattedNum = num;
  while (formattedNum >= 1000 && suffixIndex < suffixes.length - 1) {
    formattedNum /= 1000;
    suffixIndex++;
  }
  return formattedNum.toFixed(1) + suffixes[suffixIndex];
}

// Get today's traffic data
function getTodayData(dayArray) {
  if (!dayArray || dayArray.length === 0) return { rx: 0, tx: 0 };
  
  // The last element (with the largest id) in traffic.day array is today's traffic
  // Return the last element directly
  const lastIndex = dayArray.length - 1;
  return dayArray[lastIndex] || { rx: 0, tx: 0 };
}

// Fetch traffic data from server
async function fetchTrafficData() {
  const url = `${CONFIG.SERVER_URL}/json${CONFIG.TOKEN ? `?token=${CONFIG.TOKEN}` : ''}`;
  try {
    const req = new Request(url);
    req.timeoutInterval = 10;
    const response = await req.loadJSON();
    if (!response || !response.interfaces) throw new Error('Invalid response data');
    let interfaceData = response.interfaces[0];
    if (CONFIG.INTERFACE_NAME) {
      interfaceData = response.interfaces.find(iface => iface.name === CONFIG.INTERFACE_NAME) || interfaceData;
    }
    if (!interfaceData) throw new Error('Network interface data not found');
    return {
      today: getTodayData(interfaceData.traffic?.day),
      month: interfaceData.traffic?.month?.[0] || { rx: 0, tx: 0 },
      total: interfaceData.traffic?.total || { rx: 0, tx: 0 }
    };
  } catch (error) {
    console.error('Failed to fetch data:', error);
    throw error;
  }
}

// Create main widget
async function createWidget() {
  const widget = new ListWidget();
  
  try {
    const data = await fetchTrafficData();
    const fetchTime = new Date(); // Record data fetch time
    
    // Title bar (reference Cloudflare structure)
    const topStack = widget.addStack();
    topStack.topAlignContent();
    const container = topStack.addStack();
    container.layoutVertically();
    
    // Title row (no left icon, refresh time on the right)
    const titleInfo = container.addStack();
    titleInfo.layoutHorizontally();
    titleInfo.centerAlignContent();
    
    // Title (no icon)
    const title = titleInfo.addText(CONFIG.WIDGET_TITLE);
    title.font = Font.boldRoundedSystemFont(18);
    title.textColor = Color.dynamic(Color.black(), Color.white());
    
    titleInfo.addSpacer();
    
    // Refresh time (top right)
    const clockIcon = titleInfo.addImage(SFSymbol.named('clock').image);
    clockIcon.imageSize = scaleImage(clockIcon.image.size, 9);
    clockIcon.tintColor = Color.gray();
    titleInfo.addSpacer(2);
    const timeStr = fetchTime.toLocaleTimeString('en-US', { 
      hour: '2-digit', 
      minute: '2-digit',
      hour12: false
    });
    const timeLabel = titleInfo.addText(timeStr);
    timeLabel.font = Font.boldRoundedSystemFont(10);
    timeLabel.textColor = Color.gray();
    
    widget.addSpacer(10);
    
    // Today and month metrics (reference Pages and Workers layout)
    const metricStack = widget.addStack();
    metricStack.layoutHorizontally();
    
    // Today container (corresponds to Pages)
    const todayContainer = metricStack.addStack();
    todayContainer.layoutVertically();
    const todayTitleContainer = todayContainer.addStack();
    todayTitleContainer.layoutHorizontally();
    todayTitleContainer.centerAlignContent();
    
    // Today icon (using calendar icon)
    const todayIcon = todayTitleContainer.addImage(SFSymbol.named('calendar').image);
    todayIcon.imageSize = scaleImage(todayIcon.image.size, 10);
    todayIcon.tintColor = Color.orange();
    todayTitleContainer.addSpacer(2);
    const todayLabel = todayTitleContainer.addText('Today');
    todayLabel.font = Font.boldRoundedSystemFont(12);
    todayLabel.textColor = Color.gray();
    
    todayContainer.addSpacer(3);
    
    // Today value
    const todayTotal = data.today.rx + data.today.tx;
    const todayValueLabel = todayContainer.addText(formatBytes(todayTotal));
    todayValueLabel.font = Font.boldRoundedSystemFont(14);
    todayValueLabel.textColor = Color.dynamic(Color.black(), Color.white());
    
    metricStack.addSpacer(20);
    
    // Month container (corresponds to Workers)
    const monthContainer = metricStack.addStack();
    monthContainer.layoutVertically();
    const monthTitleContainer = monthContainer.addStack();
    monthTitleContainer.layoutHorizontally();
    monthTitleContainer.centerAlignContent();
    
    // Month icon (using calendar with clock badge icon)
    const monthIcon = monthTitleContainer.addImage(SFSymbol.named('calendar.badge.clock').image);
    monthIcon.imageSize = scaleImage(monthIcon.image.size, 10);
    monthIcon.tintColor = Color.orange();
    monthTitleContainer.addSpacer(2);
    const monthLabel = monthTitleContainer.addText('Month');
    monthLabel.font = Font.boldRoundedSystemFont(12);
    monthLabel.textColor = Color.gray();
    
    monthContainer.addSpacer(3);
    
    // Month value
    const monthTotal = data.month.rx + data.month.tx;
    const monthValueLabel = monthContainer.addText(formatBytes(monthTotal));
    monthValueLabel.font = Font.boldRoundedSystemFont(14);
    monthValueLabel.textColor = Color.dynamic(Color.black(), Color.white());
    
    widget.addSpacer(5);
    
    // Progress information container (corresponds to Remaining)
    const progressContainer = widget.addStack();
    progressContainer.layoutVertically();
    
    const progressTitleContainer = progressContainer.addStack();
    progressTitleContainer.layoutHorizontally();
    progressTitleContainer.centerAlignContent();
    
    // Progress icon
    const progressIcon = progressTitleContainer.addImage(SFSymbol.named('gauge').image);
    progressIcon.imageSize = scaleImage(progressIcon.image.size, 10);
    progressIcon.tintColor = Color.orange();
    progressTitleContainer.addSpacer(2);
    const progressTitleLabel = progressTitleContainer.addText('Month Used');
    progressTitleLabel.font = Font.boldRoundedSystemFont(12);
    progressTitleLabel.textColor = Color.gray();
    
    progressTitleContainer.addSpacer(10);
    
    // Display usage percentage on the right (replaces original date)
    const monthUsedGB = monthTotal / (1024 * 1024 * 1024);
    const usedPercent = Math.min((monthUsedGB / CONFIG.MONTHLY_LIMIT_GB) * 100, 100);
    const percentLabel = progressTitleContainer.addText(`${usedPercent.toFixed(1)}%`);
    percentLabel.font = Font.boldRoundedSystemFont(10);
    percentLabel.textColor = Color.dynamic(Color.black(), Color.white());
    
    progressContainer.addSpacer(3);
    
    // Progress value
    const progressValueLabel = progressContainer.addText(
      `${monthUsedGB.toFixed(1)}GB / ${CONFIG.MONTHLY_LIMIT_GB}GB`
    );
    progressValueLabel.font = Font.boldRoundedSystemFont(14);
    progressValueLabel.textColor = Color.dynamic(Color.black(), Color.white());
    
    progressContainer.addSpacer(5);
    
    // Progress bar (reference Cloudflare progress bar implementation)
    const progressBarContainer = widget.addStack();
    progressBarContainer.layoutHorizontally();
    progressBarContainer.centerAlignContent();
    
    // Calculate number of filled boxes
    const usedBoxes = (usedPercent / 100) * CONFIG.BOX_COUNT;
    const fullBoxes = Math.floor(usedBoxes); // Number of fully filled boxes
    const remainder = usedBoxes - fullBoxes; // Decimal part
    const isHalfFilled = remainder > 0.1 && remainder < 0.9; // Show half-fill if remainder is between 0.1-0.9
    
    const boxs = [];
    const boxIcons = [];
    
    // Initialize all icons
    for (let i = 0; i < CONFIG.BOX_COUNT; i++) {
      if (CONFIG.IS_SQUARE) {
        boxIcons[i] = SFSymbol.named('square').image;
      } else {
        boxIcons[i] = SFSymbol.named('rectangle.portrait').image;
      }
    }
    
    // Fill complete boxes
    for (let i = 0; i < fullBoxes && i < CONFIG.BOX_COUNT; i++) {
      if (CONFIG.IS_SQUARE) {
        boxIcons[i] = SFSymbol.named('square.fill').image;
      } else {
        boxIcons[i] = SFSymbol.named('rectangle.portrait.fill').image;
      }
    }
    
    // Add half-fill box after the last full box if needed
    if (isHalfFilled && fullBoxes < CONFIG.BOX_COUNT) {
      if (CONFIG.IS_SQUARE) {
        boxIcons[fullBoxes] = SFSymbol.named('square.lefthalf.fill').image;
      } else {
        boxIcons[fullBoxes] = SFSymbol.named('rectangle.lefthalf.fill').image;
      }
    }
    
    // Render all boxes, all using orange color
    for (let i = 0; i < CONFIG.BOX_COUNT; i++) {
      boxs[i] = progressBarContainer.addImage(boxIcons[i]);
      boxs[i].imageSize = scaleImage(boxs[i].image.size, 13);
      // All boxes use orange color
      if (i < fullBoxes || (isHalfFilled && i === fullBoxes)) {
        // Filled or half-filled boxes use full orange
        boxs[i].tintColor = Color.orange();
      } else {
        // Unfilled boxes also use orange (empty icon style distinguishes them)
        boxs[i].tintColor = Color.orange();
      }
    }
    
  } catch (error) {
    // Error display
    widget.addSpacer();
    const errorStack = widget.addStack();
    errorStack.layoutVertically();
    errorStack.centerAlignContent();
    errorStack.spacing = 3;
    
    const errorIcon = errorStack.addText('⚠️');
    errorIcon.font = Font.boldRoundedSystemFont(20);
    
    const errorText = errorStack.addText('Connection Failed');
    errorText.font = Font.boldRoundedSystemFont(15);
    errorText.textColor = Color.orange();
    errorText.centerAlignText();
    
    const errorDetail = errorStack.addText(error.message || 'Check configuration');
    errorDetail.font = Font.boldRoundedSystemFont(12);
    errorDetail.textColor = Color.gray();
    errorDetail.centerAlignText();
  }
  
  // Set refresh time
  widget.refreshAfterDate = new Date(Date.now() + CONFIG.REFRESH_INTERVAL * 1000);
  
  return widget;
}

// Execute
const widget = await createWidget();
Script.setWidget(widget);
if (!config.runsInWidget) {
  await widget.presentSmall();
}
Script.complete();
