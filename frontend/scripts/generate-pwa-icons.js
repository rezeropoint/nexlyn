/**
 * 生成PWA应用图标
 * 将SVG转换为多个尺寸的PNG文件
 *
 * 使用方法：
 * 1. 安装依赖: npm install --save-dev sharp
 * 2. 运行脚本: node scripts/generate-pwa-icons.js
 */

const sharp = require('sharp');
const fs = require('node:fs');
const path = require('node:path');

// 配置
const config = {
  inputSvg: path.join(__dirname, '../public/logos/icon-pwa.svg'),
  faviconSvg: path.join(__dirname, '../public/logos/favicon.svg'),
  outputDir: path.join(__dirname, '../public/icons'),
  sizes: [128, 192, 256, 384, 512],
  faviconSizes: [16, 32, 48, 64]
};

// 确保输出目录存在
if (!fs.existsSync(config.outputDir)) {
  fs.mkdirSync(config.outputDir, { recursive: true });
}

// 生成PWA图标
async function generatePwaIcons() {
  console.log('🎨 开始生成PWA应用图标...\n');

  try {
    // 读取SVG文件
    const svgBuffer = fs.readFileSync(config.inputSvg);

    // 生成各种尺寸的PNG
    for (const size of config.sizes) {
      const outputPath = path.join(config.outputDir, `icon-${size}x${size}.png`);

      await sharp(svgBuffer)
        .resize(size, size)
        .png()
        .toFile(outputPath);

      console.log(`✅ 生成: icon-${size}x${size}.png`);
    }

    console.log('\n🎉 PWA图标生成完成！');
    console.log(`📁 输出目录: ${config.outputDir}`);

  } catch (error) {
    console.error('❌ 生成图标失败:', error.message);
    process.exit(1);
  }
}

// 生成Favicon
async function generateFavicon() {
  console.log('\n🎨 开始生成Favicon...\n');

  try {
    const svgBuffer = fs.readFileSync(config.faviconSvg);

    // 生成多尺寸favicon PNG文件
    for (const size of config.faviconSizes) {
      const outputPath = path.join(__dirname, '../public', `favicon-${size}x${size}.png`);

      await sharp(svgBuffer)
        .resize(size, size)
        .png()
        .toFile(outputPath);

      console.log(`✅ 生成: favicon-${size}x${size}.png`);
    }

    // 生成favicon.ico（32x32单尺寸）
    const faviconIcoPath = path.join(__dirname, '../public/favicon.ico');
    await sharp(svgBuffer)
      .resize(32, 32)
      .png()
      .toFile(faviconIcoPath.replace('.ico', '-temp.png'));

    // 重命名为.ico（浏览器会识别）
    const tempPngPath = faviconIcoPath.replace('.ico', '-temp.png');
    const icoBuffer = fs.readFileSync(tempPngPath);
    fs.writeFileSync(faviconIcoPath, icoBuffer);
    fs.unlinkSync(tempPngPath);

    console.log(`✅ 生成: favicon.ico`);
    console.log('\n💡 注意: 生成的favicon.ico是PNG格式（浏览器兼容）');
    console.log('   如需真正的ICO格式，请使用在线工具: https://www.favicon-generator.org/');

  } catch (error) {
    console.error('❌ 生成Favicon失败:', error.message);
  }
}

// 执行生成
(async () => {
  await generatePwaIcons();
  await generateFavicon();

  console.log('\n📝 下一步操作:');
  console.log('1. 检查生成的图标文件');
  console.log('2. 更新 manifest.json 配置');
  console.log('3. 如需生成 favicon.ico，使用在线工具合并PNG文件');
})();
