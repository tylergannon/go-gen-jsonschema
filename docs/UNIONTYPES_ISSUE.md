# Union Types Example Issue

## Current Problem
The `examples/uniontypes` example fails with:
```
panic: resolveLocalInterfaceProps could not resolve package for type Time at time
```

## Root Cause
The example uses `time.Time` fields in structs:
- `Drawing.CreatedAt time.Time`
- `Payment.Date time.Time`

This triggers the same issue as described in `TIME_TYPE_NOTE.md` - the generator cannot handle `time.Time` from external packages.

## How to Fix the Example
1. **Short term**: Comment out or remove the `time.Time` fields to demonstrate the union type functionality
2. **Long term**: Implement special case handling for `time.Time` as described in `TIME_TYPE_NOTE.md`

## What Union Types Should Demonstrate
The example is actually well-designed to show:
- Interface-based union types (Shape, PaymentMethod)
- Multiple implementations (Circle, Rectangle, Triangle for Shape)
- Using interfaces in struct fields (Drawing.Shapes, Payment.Method)
- Both value and pointer receivers (*DigitalWallet uses pointer receiver)

Once the `time.Time` issue is resolved, this example should work properly to demonstrate union type functionality in JSON schema generation.