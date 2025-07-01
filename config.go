package main

const POLL_FUNCTION = `
(function() {
				// 设置开始时间
				const startTime = window.pollStartTime || (window.pollStartTime = Date.now());
				const elapsed = Date.now() - startTime;
				
				// 如果超过2秒，直接返回 true 继续执行
				if (elapsed > 2000) return true;
				
				// 检查标准 img 元素
				const imgs = Array.from(document.querySelectorAll('img'));
				const imgLoaded = imgs.length === 0 || imgs.every(img => 
					img.complete 
				);
				
				
				return imgLoaded;
			})()
`
