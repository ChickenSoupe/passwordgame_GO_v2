// Animation Queue Manager - Prevents ALL interruptions
class AnimationQueueManager {
    constructor() {
        this.isAnimating = false;
        this.pendingUpdates = [];
        this.animationDuration = 500; // ms
        this.latestInputValue = null; // Track latest input value
    }

    // Queue an update to be processed when animations finish
    queueUpdate(updateFunction) {
        if (this.isAnimating) {
            // Replace any pending updates with this latest one
            this.pendingUpdates = [updateFunction];
            console.log('Animation in progress - queuing update');
        } else {
            // Execute immediately if not animating
            updateFunction();
        }
    }

    // Process the queue if not currently animating
    processQueue() {
        if (this.isAnimating || this.pendingUpdates.length === 0) {
            return;
        }

        console.log('Processing queued updates');
        // Execute all pending updates
        while (this.pendingUpdates.length > 0) {
            const update = this.pendingUpdates.shift();
            update();
        }
    }

    // Mark animation as started
    startAnimation() {
        this.isAnimating = true;
        console.log('Animation started - blocking all DOM updates');
        
        // Auto-clear after animation duration + buffer
        setTimeout(() => {
            this.finishAnimation();
        }, this.animationDuration + 100);
    }

    // Mark animation as finished and process queue
    finishAnimation() {
        this.isAnimating = false;
        console.log('Animation finished - processing queued updates');
        
        // Process any queued updates
        requestAnimationFrame(() => {
            this.processQueue();
        });
    }

    // Force finish (emergency cleanup)
    forceFinish() {
        this.isAnimating = false;
        this.pendingUpdates = [];
        
        // Clean up any existing animations
        const animatingRules = document.querySelectorAll('.flip-animate');
        animatingRules.forEach(rule => {
            rule.style.transform = '';
            rule.style.transition = '';
            rule.classList.remove('flip-animate');
        });
    }

    // Check if currently animating
    isCurrentlyAnimating() {
        return this.isAnimating;
    }
}

// Enhanced Smart Debouncer with queue integration
class SmartDebouncer {
    constructor(delay = 300, animationQueue) {
        this.delay = delay;
        this.timeoutId = null;
        this.callback = null;
        this.pendingValue = null;
        this.animationQueue = animationQueue;
    }

    debounce(callback, value) {
        this.callback = callback;
        this.pendingValue = value;

        // Clear existing timeout
        if (this.timeoutId) {
            clearTimeout(this.timeoutId);
        }

        // Start debounce timer
        this.timeoutId = setTimeout(() => {
            this.execute();
        }, this.delay);
    }

    execute() {
        if (this.callback && this.pendingValue !== null) {
            const callback = this.callback;
            const value = this.pendingValue;
            
            // Queue the update through animation manager
            this.animationQueue.queueUpdate(() => {
                callback(value);
            });
            
            this.pendingValue = null;
        }
    }

    cancel() {
        if (this.timeoutId) {
            clearTimeout(this.timeoutId);
            this.timeoutId = null;
        }
    }
}

// Rule State Manager
class RuleStateManager {
    constructor() {
        this.previousStates = {
            satisfied: {},
            visible: {}
        };
        this.currentStates = {
            satisfied: {},
            visible: {}
        };
    }

    updateStates(satisfiedStates, visibleStates) {
        this.previousStates = {
            satisfied: { ...this.currentStates.satisfied },
            visible: { ...this.currentStates.visible }
        };

        this.currentStates = {
            satisfied: { ...satisfiedStates },
            visible: { ...visibleStates }
        };
    }

    hasRuleStateChanges() {
        const satisfactionChanged = this.hasStateChanged(
            this.previousStates.satisfied,
            this.currentStates.satisfied
        );

        const visibilityChanged = this.hasStateChanged(
            this.previousStates.visible,
            this.currentStates.visible
        );

        return satisfactionChanged || visibilityChanged;
    }

    hasStateChanged(previous, current) {
        const prevKeys = Object.keys(previous);
        const currKeys = Object.keys(current);

        if (prevKeys.length !== currKeys.length) {
            return true;
        }

        for (const key of currKeys) {
            if (previous[key] !== current[key]) {
                return true;
            }
        }

        return false;
    }
}

// Enhanced FLIP Animator with queue integration
class FLIPAnimator {
    constructor(ruleStateManager, animationQueue) {
        this.firstStates = new Map();
        this.ruleStateManager = ruleStateManager;
        this.animationQueue = animationQueue;
        this.animationDuration = 500; // ms
    }

    recordFirst() {
        const rules = document.querySelectorAll('.rule-item:not(.initially-hidden)');
        this.firstStates.clear();
        
        rules.forEach(rule => {
            const rect = rule.getBoundingClientRect();
            this.firstStates.set(rule.dataset.ruleId, {
                top: rect.top,
                left: rect.left,
                width: rect.width,
                height: rect.height
            });
        });
    }

    shouldAnimate() {
        return this.ruleStateManager.hasRuleStateChanges();
    }

    animateToLast() {
        if (!this.shouldAnimate()) {
            console.log('No rule state changes - skipping animation');
            return;
        }

        if (this.firstStates.size === 0) {
            console.log('No first states recorded - skipping animation');
            return;
        }

        const rules = document.querySelectorAll('.rule-item:not(.initially-hidden)');
        const animatingRules = [];

        rules.forEach(rule => {
            const ruleId = rule.dataset.ruleId;
            const first = this.firstStates.get(ruleId);
            
            if (!first) return;

            const last = rule.getBoundingClientRect();
            const deltaX = first.left - last.left;
            const deltaY = first.top - last.top;

            if (Math.abs(deltaX) > 1 || Math.abs(deltaY) > 1) {
                rule.style.transform = `translate(${deltaX}px, ${deltaY}px)`;
                rule.style.transition = 'none';
                
                rule.offsetHeight; // Force reflow
                
                rule.classList.add('flip-animate');
                rule.style.transition = `transform ${this.animationDuration}ms cubic-bezier(0.25, 0.46, 0.45, 0.94)`;
                rule.style.transform = 'translate(0, 0)';
                
                animatingRules.push(rule);
            }
        });

        if (animatingRules.length > 0) {
            this.animationQueue.startAnimation();
            console.log(`Animating ${animatingRules.length} rules`);
            
            const cleanup = () => {
                animatingRules.forEach(rule => {
                    rule.style.transform = '';
                    rule.style.transition = '';
                    rule.classList.remove('flip-animate');
                });
                this.animationQueue.finishAnimation();
                console.log('Animation cleanup completed');
            };

            // Primary cleanup mechanism
            setTimeout(cleanup, this.animationDuration);
            
            // Backup cleanup
            if (animatingRules[0]) {
                animatingRules[0].addEventListener('transitionend', cleanup, { once: true });
            }
        } else {
            console.log('No visual movement - no animation needed');
        }
    }
}

// Initialize and export for use in other scripts
window.AnimationSystem = {
    AnimationQueueManager,
    SmartDebouncer,
    RuleStateManager,
    FLIPAnimator
};