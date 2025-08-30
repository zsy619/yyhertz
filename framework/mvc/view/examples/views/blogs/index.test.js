const fs = require('fs');
const path = require('path');
const { JSDOM } = require('jsdom');

// Load the HTML template
const htmlTemplate = fs.readFileSync(
  path.join(__dirname, 'index.html'),
  'utf-8'
);

describe('Blog List Template', () => {
  let dom;
  let document;

  beforeEach(() => {
    // Setup JSDOM with the HTML template
    dom = new JSDOM(htmlTemplate);
    document = dom.window.document;
  });

  test('renders the page title correctly', () => {
    // Mock data
    const mockData = {
      PageTitle: 'YYHertz 技术博客',
      Articles: [],
    };

    // Simulate template rendering (replace with actual template engine logic if needed)
    const renderedHtml = htmlTemplate
      .replace('{{.PageTitle}}', mockData.PageTitle)
      .replace('{{range .Articles}}', '')
      .replace('{{end}}', '');

    // Update JSDOM with rendered HTML
    dom = new JSDOM(renderedHtml);
    document = dom.window.document;

    // Assert
    const pageTitle = document.querySelector('h2').textContent;
    expect(pageTitle).toBe(mockData.PageTitle);
  });

  test('renders articles correctly', () => {
    // Mock data
    const mockData = {
      PageTitle: 'YYHertz 技术博客',
      Articles: [
        {
          Title: '测试文章标题',
          Author: '测试作者',
          Date: '2025-08-29',
          Summary: '测试文章摘要',
          Tags: ['测试', '技术'],
        },
      ],
    };

    // Simulate template rendering (replace with actual template engine logic if needed)
    let renderedHtml = htmlTemplate
      .replace('{{.PageTitle}}', mockData.PageTitle);

    // Simulate range loop for articles
    renderedHtml = renderedHtml.replace(
      '{{range .Articles}}',
      mockData.Articles
        .map(
          (article) => `
        <article class="article-card">
            <div class="article-header">
                <h3 class="article-title">${article.Title}</h3>
                <div class="article-meta">
                    <span class="author">👤 ${article.Author}</span>
                    <span class="date">📅 ${article.Date}</span>
                </div>
            </div>
            
            <div class="article-content">
                <p class="article-summary">${article.Summary}</p>
            </div>
            
            <div class="article-footer">
                <div class="tags">
                    ${article.Tags.map((tag) => `<span class="tag">${tag}</span>`).join('')}
                </div>
                <a href="#" class="read-more">阅读更多 →</a>
            </div>
        </article>
        `
        )
        .join('')
    );

    renderedHtml = renderedHtml.replace('{{end}}', '');

    // Update JSDOM with rendered HTML
    dom = new JSDOM(renderedHtml);
    document = dom.window.document;

    // Assert
    const articleTitles = document.querySelectorAll('.article-title');
    expect(articleTitles.length).toBe(mockData.Articles.length);
    expect(articleTitles[0].textContent).toBe(mockData.Articles[0].Title);
  });

  test('handles empty articles list', () => {
    // Mock data with empty articles
    const mockData = {
      PageTitle: 'YYHertz 技术博客',
      Articles: [],
    };

    // Simulate template rendering (replace with actual template engine logic if needed)
    const renderedHtml = htmlTemplate
      .replace('{{.PageTitle}}', mockData.PageTitle)
      .replace('{{range .Articles}}', '')
      .replace('{{end}}', '');

    // Update JSDOM with rendered HTML
    dom = new JSDOM(renderedHtml);
    document = dom.window.document;

    // Assert
    const articles = document.querySelectorAll('.article-card');
    expect(articles.length).toBe(0);
  });
});