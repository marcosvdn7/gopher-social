import { useState } from "react"

export const CreatePostForm: React.FC = () => {
    const [title, setTitle] = useState('')
    const [content, setContent] = useState('')

    const handleSubmit = async () => {
        await fetch(`http://localhost:8081/v1/posts`, {
            method: "POST",
            headers: {
                "Content-Type": "application/json",
                Authorization: `Bearer foo`
            },
            body: JSON.stringify({
                title,
                content
            })
        })

        setTitle('')
        setContent('')
    }

    return (
        <div className="gopher-social">
            <label>
                <input placeholder="Title..." value={title} type="text" 
                onChange={(e) => setTitle(e.target.value)}></input>
            </label>
            <label>
                <textarea placeholder="What's on you mind..." value={content} 
                onChange={(e) => setContent(e.target.value)}></textarea>
            </label>
            <button onClick={handleSubmit}>Share</button>
        </div>
    )
}
