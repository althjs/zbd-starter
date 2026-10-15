<script lang="ts">
	/** 스타터 로고 — 즈비드 글자 로고(획) + 하트 + 「Starter」. 관리자 머리줄·로그인 화면·사이트 머리 공용 */
	let { height = 28, color = '#fff', heart = '#F87171' }: { height?: number; color?: string; heart?: string } = $props();
</script>

<svg viewBox="-12 -52 1230 262" style:height="{height}px" style:width="auto" role="img" aria-label="ZBD Starter" xmlns="http://www.w3.org/2000/svg">
	<g fill="none" style:stroke={color} stroke-width="40" stroke-linecap="round" stroke-linejoin="round">
		<path d="M22,22 H128" />
		<path d="M75,30 C75,70 48,96 20,114" />
		<path d="M75,30 C75,70 102,96 130,114" />
		<path d="M10,180 H140" />
		<path d="M202,14 V108 Q202,130 224,130 H266 Q288,130 288,108 V14" />
		<path d="M202,64 H288" />
		<path d="M348,10 V186" />
		<path d="M528,22 H414 V124 H528" />
		<path d="M398,180 H542" />
	</g>
	<path style:fill={heart} d="M600,62 C572,40 548,20 548,-2 C548,-18 560,-30 575,-30 C586,-30 595,-24 600,-15 C605,-24 614,-30 625,-30 C640,-30 652,-18 652,-2 C652,20 628,40 600,62 Z" />
	<text x="690" y="178" style:fill={color} font-family="system-ui, -apple-system, 'Segoe UI', sans-serif" font-size="150" font-weight="800" letter-spacing="-2">Starter</text>
</svg>
