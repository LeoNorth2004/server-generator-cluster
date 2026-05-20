export function GlassCard({ children, className = '', hover = false, ...props }) {
  return (
    <div
      className={`${hover ? 'glass-card-hover' : 'glass-card'} p-6 ${className}`}
      {...props}
    >
      {children}
    </div>
  );
}

export function StatCard({ title, value, icon, trend, subtitle }) {
  const isNumericTrend = typeof trend === 'number' || (typeof trend === 'string' && !isNaN(trend) && trend !== '');

  return (
    <GlassCard hover>
      <div className="flex items-start justify-between">
        <div className="flex flex-col gap-1 min-w-0 flex-1">
          <span className="stat-label">{title}</span>
          <span className="stat-value break-all overflow-wrap-anywhere">{value}</span>
          {isNumericTrend && (
            <span className={`text-sm ${Number(trend) > 0 ? 'text-green-500' : 'text-red-500'}`}>
              {Number(trend) > 0 ? '+' : ''}{trend}%
            </span>
          )}
          {subtitle && !isNumericTrend && (
            <span className="text-sm text-gray-500 dark:text-gray-400">
              {subtitle}
            </span>
          )}
        </div>
        {icon && (
          <div className="p-3 rounded-xl bg-primary-50 dark:bg-primary-950/50 text-primary-600 dark:text-primary-400 flex-shrink-0 ml-2">
            {icon}
          </div>
        )}
      </div>
    </GlassCard>
  );
}

export function ChartCard({ title, children, action }) {
  return (
    <GlassCard className="h-full">
      <div className="flex items-center justify-between mb-4">
        <h3 className="text-lg font-semibold text-gray-900 dark:text-white">{title}</h3>
        {action}
      </div>
      {children}
    </GlassCard>
  );
}

export default { GlassCard, StatCard, ChartCard };