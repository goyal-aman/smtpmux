def selector(downstreams):
    print("Executing Starlark selector...")
    
    # Logic: Try the first one, if it fails, try the second
    for ds in downstreams:
        print("Attempting to send via:", ds["addr"])
        err = send(ds=ds)
        if err == None:
            return None
        print("Failed:", err)
    
    return "all downstreams failed"
