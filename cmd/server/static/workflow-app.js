function injectNavbar(currentPage) {
    const routes = { '/': 'Home', '/payment': 'Payment', '/order': 'Order', '/failing': 'Failing' };
    const active = currentPage || window.location.pathname;
    const html = Object.entries(routes).map(([path, label]) => {
        const isActive = active === path;
        return `<a href="${path}" class="px-4 py-2 text-sm border-b-2 transition ${isActive ? 'text-gray-800 border-gray-800' : 'text-gray-400 border-transparent hover:text-gray-600 hover:border-gray-300'}">${label}</a>`;
    }).join('');
    document.body.insertAdjacentHTML('afterbegin', `<nav class="flex justify-center gap-2 mb-8">${html}</nav>`);
}

function workflowApp(config) {
    return {
        workflowId: '',
        progress: 0,
        activity: '',
        loading: false,
        error: '',
        copied: false,
        pollInterval: null,
        timeline: [],
        timelineTotalMs: 0,
        timelineStartMs: 0,

        init() {
            this.fields = config.fields;
            config.fields.forEach(f => this[f.model] = f.default || '');
        },

        startWorkflow() {
            this.loading = true;
            this.error = '';
            this.progress = 0;
            this.timeline = [];

            const payload = {};
            config.fields.forEach(f => {
                let val = this[f.model];
                if (f.transform) val = f.transform(val);
                payload[f.key] = f.type === 'number' ? parseFloat(val) : val;
            });

            fetch(config.apiEndpoint, {
                method: 'POST',
                headers: { 'Content-Type': 'application/json' },
                body: JSON.stringify(payload)
            })
            .then(res => res.json())
            .then(data => {
                this.workflowId = data.workflow_id;
                this.loading = false;
                this.startPolling();
            })
            .catch(err => {
                this.error = 'Failed to start: ' + err.message;
                this.loading = false;
            });
        },

        startPolling() {
            this.pollInterval = setInterval(() => this.poll(), 1000);
        },

        poll() {
            const url = `/api/workflow/timeline?workflow_id=${this.workflowId}` +
                (config.expectedTotal ? '&expected_total=true' : '');
            fetch(url)
                .then(r => r.json())
                .then(tl => {
                    if (config.onTimeline) {
                        config.onTimeline.call(this, tl);
                    } else {
                        this.handleTimeline(tl);
                    }
                })
                .catch(err => console.error('Poll error:', err));
        },

        handleTimeline(tl) {
            this.progress = tl.progress || 0;
            if (tl.activities) {
                this.timeline = tl.activities;
                this.timelineStartMs = tl.started_at_ms || 0;
                this.timelineTotalMs = (tl.ended_at_ms || Date.now()) - this.timelineStartMs;

                const failed = tl.activities.find(a => a.status === 'failed');
                if (failed) {
                    const completed = tl.activities.filter(a => a.status === 'completed').length;
                    this.progress = tl.total_activities > 0 ? Math.round((completed / tl.total_activities) * 100) : 0;
                    this.error = 'Failed at: ' + failed.name;
                    this.activity = failed.name;
                    this.stopPolling();
                    return;
                }

                const last = tl.activities[tl.activities.length - 1];
                if (last) this.activity = last.name;
            }
            if (this.progress == 0) this.activity = 'Starting...';
            if (tl.ended_at_ms || tl.progress >= 100) {
                this.progress = 100;
                this.activity = 'Complete';
                this.stopPolling();
            }
        },

        stopPolling() {
            if (this.pollInterval) {
                clearInterval(this.pollInterval);
                this.pollInterval = null;
            }
        },

        timelineBarStyle(span) {
            if (!this.timelineStartMs || !this.timelineTotalMs) return 'left:0;width:100%';
            const left = ((span.started_at_ms - this.timelineStartMs) / this.timelineTotalMs) * 100;
            let width = span.ended_at_ms
                ? ((span.ended_at_ms - span.started_at_ms) / this.timelineTotalMs) * 100
                : ((Date.now() - span.started_at_ms) / this.timelineTotalMs) * 100;
            return `left:${Math.max(0, left).toFixed(1)}%;width:${Math.min(100 - left, width).toFixed(1)}%`;
        },

        copyWorkflowId() {
            try {
                navigator.clipboard.writeText(this.workflowId);
                this.copied = true;
                setTimeout(() => this.copied = false, 2000);
            } catch (e) {
                console.warn('Clipboard unavailable:', e);
            }
        },

        reset() {
            this.workflowId = '';
            this.progress = 0;
            this.activity = '';
            this.timeline = [];
            this.error = '';
            this.stopPolling();
        }
    };
}