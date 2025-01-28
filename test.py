import asyncio
async def cancel_me():
    print('cancel_me(): before sleep')

    try:
        # Wait for 1 hour
        await asyncio.sleep(3600)
    except asyncio.CancelledError:
        print('cancel_me(): cancel sleep')
        # raise
    finally:
        print('cancel_me(): after sleep')

async def main():
    # Create a "cancel_me" Task
    task = asyncio.create_task(cancel_me())

    # Wait for 1 second
    await asyncio.sleep(1)

    print('main(): before cancel')
    for t in asyncio.all_tasks():
        print(t)
    task.cancel()
    print('main(): after cancel')
    for t in asyncio.all_tasks():
        print(t)

    try:
        print('main(): before await task')
        for t in asyncio.all_tasks():
            print(t)
        await task
        print('main(): after await task')
        for t in asyncio.all_tasks():
            print(t)
    except asyncio.CancelledError:
        print("main(): cancel_me is cancelled now")

    print('main(): end')
    for t in asyncio.all_tasks():
        print(t)

asyncio.run(main())