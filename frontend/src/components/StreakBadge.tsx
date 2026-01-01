import { useState, useEffect, useRef } from 'react';

interface StreakBadgeProps {
  currentStreak: number;
  longestStreak?: number;
  lastSolvedDate?: string | null;
  size?: 'sm' | 'md' | 'lg';
  showTooltip?: boolean;
  animate?: boolean;
}

/**
 * Format relative date for tooltip
 */
function formatLastSolved(dateString: string | null | undefined): string {
  if (!dateString) return 'No problems solved yet';

  const date = new Date(dateString);
  const now = new Date();
  const diffMs = now.getTime() - date.getTime();
  const diffDays = Math.floor(diffMs / (1000 * 60 * 60 * 24));

  if (diffDays === 0) return 'Solved today';
  if (diffDays === 1) return 'Solved yesterday';
  if (diffDays < 7) return `Solved ${diffDays} days ago`;

  return date.toLocaleDateString('en-US', {
    month: 'short',
    day: 'numeric',
    year: date.getFullYear() !== now.getFullYear() ? 'numeric' : undefined,
  });
}

type FlameColor = 'grayscale' | 'orange' | 'red' | 'purple';

/**
 * Get flame color based on streak length
 */
function getFlameColor(streak: number): FlameColor {
  if (streak === 0) return 'grayscale';
  if (streak < 7) return 'orange';
  if (streak < 30) return 'red';
  return 'purple'; // Legendary streak
}

/**
 * Get size classes
 */
function getSizeClasses(size: 'sm' | 'md' | 'lg'): { container: string; flame: string; text: string } {
  switch (size) {
    case 'sm':
      return { container: 'px-2 py-1', flame: 'text-lg', text: 'text-sm' };
    case 'lg':
      return { container: 'px-4 py-2', flame: 'text-3xl', text: 'text-xl' };
    default:
      return { container: 'px-3 py-1.5', flame: 'text-2xl', text: 'text-base' };
  }
}

/**
 * StreakBadge - Displays current streak with animated flame
 */
export default function StreakBadge({
  currentStreak,
  longestStreak,
  lastSolvedDate,
  size = 'md',
  showTooltip = true,
  animate = true,
}: StreakBadgeProps) {
  const [isAnimating, setIsAnimating] = useState(false);
  const [showTooltipState, setShowTooltipState] = useState(false);
  const prevStreakRef = useRef(currentStreak);
  const tooltipRef = useRef<HTMLDivElement>(null);

  // Trigger animation when streak increases
  useEffect(() => {
    if (currentStreak > prevStreakRef.current && animate) {
      setIsAnimating(true);
      const timer = setTimeout(() => setIsAnimating(false), 1000);
      return () => clearTimeout(timer);
    }
    prevStreakRef.current = currentStreak;
  }, [currentStreak, animate]);

  const flameColor = getFlameColor(currentStreak);
  const sizeClasses = getSizeClasses(size);
  const isActive = currentStreak > 0;

  // Color classes based on streak
  const colorClasses = {
    grayscale: 'bg-gray-800 border-gray-600',
    orange: 'bg-orange-900/40 border-orange-600',
    red: 'bg-red-900/40 border-red-600',
    purple: 'bg-purple-900/40 border-purple-600',
  };

  const textColorClasses = {
    grayscale: 'text-gray-400',
    orange: 'text-orange-300',
    red: 'text-red-300',
    purple: 'text-purple-300',
  };

  return (
    <div
      className="relative inline-block"
      onMouseEnter={() => showTooltip && setShowTooltipState(true)}
      onMouseLeave={() => setShowTooltipState(false)}
    >
      {/* Badge */}
      <div
        className={`
          flex items-center gap-2 rounded-full border transition-all duration-300
          ${sizeClasses.container}
          ${colorClasses[flameColor]}
          ${isAnimating ? 'scale-110 shadow-lg shadow-orange-500/50' : ''}
        `}
      >
        {/* Flame emoji with animation */}
        <span
          className={`
            ${sizeClasses.flame}
            ${isActive ? '' : 'grayscale opacity-50'}
            ${isAnimating ? 'animate-bounce' : isActive && animate ? 'animate-pulse' : ''}
          `}
          role="img"
          aria-label="streak flame"
        >
          {currentStreak >= 30 ? '🔥' : currentStreak >= 7 ? '🔥' : currentStreak > 0 ? '🔥' : '💨'}
        </span>

        {/* Streak count */}
        <span className={`font-bold ${sizeClasses.text} ${textColorClasses[flameColor]}`}>
          {currentStreak}
        </span>

        {/* Day label for larger sizes */}
        {size !== 'sm' && (
          <span className={`text-gray-400 ${size === 'lg' ? 'text-base' : 'text-sm'}`}>
            {currentStreak === 1 ? 'day' : 'days'}
          </span>
        )}
      </div>

      {/* Tooltip */}
      {showTooltip && showTooltipState && (
        <div
          ref={tooltipRef}
          className="absolute z-50 bottom-full left-1/2 -translate-x-1/2 mb-2 px-3 py-2 bg-gray-900 border border-gray-700 rounded-lg shadow-xl min-w-[180px]"
        >
          <div className="text-center">
            <p className="text-white font-semibold mb-1">
              {currentStreak === 0 ? 'No Active Streak' : `${currentStreak} Day Streak!`}
            </p>
            <p className="text-gray-400 text-sm">
              {formatLastSolved(lastSolvedDate)}
            </p>
            {longestStreak !== undefined && longestStreak > 0 && (
              <p className="text-gray-500 text-xs mt-1">
                Best: {longestStreak} days
              </p>
            )}
          </div>
          {/* Tooltip arrow */}
          <div className="absolute top-full left-1/2 -translate-x-1/2 -mt-px">
            <div className="border-8 border-transparent border-t-gray-700" />
            <div className="absolute top-0 left-1/2 -translate-x-1/2 border-[7px] border-transparent border-t-gray-900" />
          </div>
        </div>
      )}
    </div>
  );
}

/**
 * Compact streak display for inline use
 */
export function StreakBadgeCompact({ currentStreak }: { currentStreak: number }) {
  const isActive = currentStreak > 0;

  return (
    <span className={`inline-flex items-center gap-1 ${isActive ? 'text-orange-400' : 'text-gray-500'}`}>
      <span className={isActive ? '' : 'grayscale opacity-50'}>🔥</span>
      <span className="font-mono font-bold">{currentStreak}</span>
    </span>
  );
}
