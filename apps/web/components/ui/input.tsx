import { forwardRef, InputHTMLAttributes } from "react";
import { cn } from "@/lib/utils";

interface InputProps extends InputHTMLAttributes<HTMLInputElement> {
  label?: string;
  error?: string;
}

export const Input = forwardRef<HTMLInputElement, InputProps>(
  ({ className, label, error, id, ...rest }, ref) => {
    const inputId = id || label?.toLowerCase().replace(/\s+/g, "-");
    // Only coalesce when the caller explicitly opted into a controlled `value` — if it was
    // never passed at all (e.g. react-hook-form's uncontrolled register()), leave the input
    // alone. File inputs can never take a `value`. This stops a value that's momentarily
    // undefined (e.g. before async data loads) from flipping the input from controlled to
    // uncontrolled, which React warns about and can drop keystrokes over.
    if ("value" in rest && rest.type !== "file" && rest.value === undefined) {
      rest.value = "";
    }
    return (
      <div className="space-y-1.5">
        {label && (
          <label htmlFor={inputId} className="block text-sm font-medium text-slate-700">
            {label}
          </label>
        )}
        <input
          ref={ref}
          id={inputId}
          className={cn(
            "block w-full rounded-lg border border-slate-300 px-3 py-2 text-slate-900 placeholder-slate-400 shadow-sm transition-colors focus:border-brand-500 focus:outline-none focus:ring-1 focus:ring-brand-500",
            error && "border-red-500 focus:border-red-500 focus:ring-red-500",
            className,
          )}
          {...rest}
        />
        {error && <p className="text-sm text-red-600">{error}</p>}
      </div>
    );
  },
);
Input.displayName = "Input";
