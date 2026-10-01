from archon import agent, vector, llm


@agent.handler
async def handle(req):
    query = req.json["message"]
    hits = await vector("docs-index").query(text=query, top_k=5)
    context = "\n".join(h.payload["text"] for h in hits)
    return await llm("chat-llm").chat(
        [
            {"role": "system", "content": context},
            {"role": "user", "content": query},
        ],
        stream=True,
    )
